using Microsoft.EntityFrameworkCore;
using System.Linq.Expressions;

namespace API.DataAccess;

/// <summary>
/// Allows a response type to define a static projection expression for a domain type.
/// This in turn allows repositories to project domain types to response types at the database level 
/// without needing to know about the response types.
/// </summary>
/// <typeparam name="TDomain"></typeparam>
/// <typeparam name="TResponse"></typeparam>
public interface IProjectable<TDomain, TResponse>
    where TDomain : class
    where TResponse : class, IProjectable<TDomain, TResponse>
{
    /// <summary>
    /// The expression used to project from <typeparamref name="TDomain"/> to <typeparamref name="TResponse"/>.
    /// Used by repositories to perform projections at the database level to ensure only necessary data is retrieved.
    /// </summary>
    static abstract Expression<Func<TDomain, TResponse>> Projection { get; }
    public static virtual Func<TDomain, TResponse> ConvertFunction
        => field ??= TResponse.Projection.Compile();
}

// Extension method for LINQ to apply a projection expression to an IQueryable source.
public static class QueryableExtensions
{
    extension<TDomain, TResponse>(IQueryable<TDomain> source)
        where TDomain : class
        where TResponse : class, IProjectable<TDomain, TResponse>
    {
        public IQueryable<TResponse> ProjectTo()
            => source.Select(TResponse.Projection);
    }

    extension<TDomain, TResponse>(IEnumerable<TDomain> source)
        where TDomain : class
        where TResponse : class, IProjectable<TDomain, TResponse>
    {
        public IQueryable<TResponse> ProjectTo()
            => source.AsQueryable().Select(TResponse.Projection);
    }
}

public static class MappableExtensions
{
    /// <summary>
    /// Constructs a single TResponse object from a TDomain object using the static Constructor func.
    /// Used for in-memory mapping after data has been retrieved from the database.
    /// </summary>
    public static TResponse ConvertToResponse<TDomain, TResponse>(
        this TDomain source)
        where TDomain : class
        where TResponse : class, IProjectable<TDomain, TResponse>
    {
        // Executes the pre-compiled static constructor delegate
        return TResponse.ConvertFunction.Invoke(source);
    }

    /// <summary>
    /// Constructs a sequence of TResponse objects from a sequence of TDomain objects.
    /// </summary>
    public static IEnumerable<TResponse> ConvertToResponse<TDomain, TResponse>(
        this IEnumerable<TDomain> source)
        where TDomain : class
        where TResponse : class, IProjectable<TDomain, TResponse>
    {
        return source.Select(TResponse.ConvertFunction);
    }
}