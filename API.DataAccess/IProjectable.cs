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
    where TResponse : IProjectable<TDomain, TResponse>
{
    /// <summary>
    /// The expression used to project from <typeparamref name="TDomain"/> to <typeparamref name="TResponse"/>.
    /// Used by repositories to perform projections at the database level to ensure only necessary data is retrieved.
    /// </summary>
    static abstract Expression<Func<TDomain, TResponse>> Projection { get; }
}

// Extension method for LINQ to apply a projection expression to an IQueryable source.
public static class QueryableExtensions
{
    public static IQueryable<TResponse> ProjectTo<TDomain, TResponse>(
        this IQueryable<TDomain> source)
        where TDomain : class
        where TResponse : IProjectable<TDomain, TResponse>
    {
        return source.Select(TResponse.Projection);
    }  
    
    public static IQueryable<TResponse> ProjectTo<TDomain, TResponse>(
        this IEnumerable<TDomain> source)
        where TDomain : class
        where TResponse : IProjectable<TDomain, TResponse>
    {
        return source.AsQueryable().Select(TResponse.Projection);
    }

    public static IQueryable<TResponse> ProjectTo<TDomain, TResponse>(
        this TDomain source)
        where TDomain : class
        where TResponse : IProjectable<TDomain, TResponse>
    {
        return new[] { source }
            .AsQueryable()
            .Select(TResponse.Projection);
    }
}
