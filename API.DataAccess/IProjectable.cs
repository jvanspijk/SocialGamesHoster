using System.Linq.Expressions;

namespace API.DataAccess;

/// <summary>
/// Allows a response type to define a static projection expression for an entity type.
/// This in turn allows repositories to project entities to response types at the database level 
/// without needing to know about the response types.
/// </summary>
/// <typeparam name="TEntity"></typeparam>
/// <typeparam name="TProjection"></typeparam>
public interface IProjectable<TEntity, TProjection>
    where TEntity : class
    where TProjection : IProjectable<TEntity, TProjection>
{
    static abstract Expression<Func<TEntity, TProjection>> Projection { get; }
}
