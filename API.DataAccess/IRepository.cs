using API.Domain;
using System.Linq.Expressions;

namespace API.DataAccess;
public interface IRepository<TEntity> where TEntity : class, IEntity
{
    Task<TEntity?> GetReadOnlyAsync(Expression<Func<TEntity, bool>> predicate);
    Task<TResult?> GetReadOnlyAsync<TResult>(Expression<Func<TEntity, bool>> predicate)
        where TResult : class, IProjectable<TEntity, TResult>;
    Task<TEntity[]> GetArrayReadOnlyAsync();
    Task<TEntity[]> GetArrayReadOnlyAsync(Expression<Func<TEntity, bool>> predicate);
    Task<TResult[]> GetArrayReadOnlyAsync<TResult>()
    where TResult : class, IProjectable<TEntity, TResult>;

    Task<TResult[]> GetArrayReadOnlyAsync<TResult>(Expression<Func<TEntity, bool>> predicate) 
        where TResult : class, IProjectable<TEntity, TResult>;

    Task<TResult[]> QueryReadOnlyAsync<TResult>(Func<IQueryable<TEntity>, IQueryable<TEntity>> queryBuilder) 
        where TResult : class, IProjectable<TEntity, TResult>;

    Task<bool> ExistsAsync(int id);
    Task<bool> ExistsAsync(Expression<Func<TEntity, bool>> predicate);
    Task<int> CountAsync();
    Task<int> CountAsync(Expression<Func<TEntity, bool>> predicate);

    ValueTask<TEntity?> GetWithTrackingAsync(int id);
    Task<List<TEntity>> GetListWithTrackingAsync(IReadOnlyCollection<int> ids);
    Task<List<TEntity>> GetListWithTrackingAsync(Expression<Func<TEntity, bool>> predicate);

    TEntity Add(TEntity entity);
    IReadOnlyCollection<TEntity> AddMultiple(IReadOnlyCollection<TEntity> entities);

    void Remove(TEntity entity);
    void RemoveMultiple(IReadOnlyCollection<TEntity> entities);

    Task SaveChangesAsync();
    void CancelChanges();
}