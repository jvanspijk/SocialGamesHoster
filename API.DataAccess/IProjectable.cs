using System.Linq.Expressions;

namespace API.DataAccess;

/// <summary>
/// Allows a response type to define a static projection expression for an entity type.
/// This in turn allows repositories to project entities to response types at the database level 
/// without needing to know about the response types.
/// </summary>
/// <typeparam name="TDomainModel"></typeparam>
/// <typeparam name="TResponseModel"></typeparam>
public interface IProjectable<TDomainModel, TResponseModel>
    where TDomainModel : class
    where TResponseModel : IProjectable<TDomainModel, TResponseModel>
{
    static abstract Expression<Func<TDomainModel, TResponseModel>> Projection { get; }
}
