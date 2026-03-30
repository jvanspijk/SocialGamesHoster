using Microsoft.EntityFrameworkCore;
using API.Domain;
using System.Linq.Expressions;
using Microsoft.Extensions.Logging;

namespace API.DataAccess;

public class Repository<TEntity>(APIDatabaseContext context) : IRepository<TEntity> where TEntity : class, IEntity
{
    internal readonly APIDatabaseContext context = context;
    private readonly DbSet<TEntity> _dbSet = context.Set<TEntity>();

    public Task<TEntity?> GetReadOnlyAsync(Expression<Func<TEntity, bool>> predicate)
    {
        return _dbSet
            .AsNoTracking()
            .Where(predicate)
            .FirstOrDefaultAsync();
    }

    public Task<TResult?> GetReadOnlyAsync<TResult>(Expression<Func<TEntity, bool>> predicate, bool splitQuery)
        where TResult : class, IProjectable<TEntity, TResult>
    {
        var query = _dbSet
            .AsNoTracking()
            .Where(predicate)
            .Select(TResult.Projection);

        if(splitQuery)
        {
            query = query.AsSplitQuery();
        }
  
        return query.FirstOrDefaultAsync();
    }

    public Task<TResult[]> GetArrayReadOnlyAsync<TResult>()
where TResult : class, IProjectable<TEntity, TResult>
    {
        return _dbSet
            .AsNoTracking()
            .Select(TResult.Projection)
            .ToArrayAsync();
    }

    public Task<TEntity[]> GetArrayReadOnlyAsync()
    {
        return _dbSet
            .AsNoTracking()
            .ToArrayAsync();
    }

    public Task<TEntity[]> GetArrayReadOnlyAsync(Expression<Func<TEntity, bool>> predicate)
    {
        return _dbSet
            .AsNoTracking()
            .Where(predicate)
            .ToArrayAsync();
    }

    public Task<TResult[]> GetArrayReadOnlyAsync<TResult>(Expression<Func<TEntity, bool>> predicate)
        where TResult : class, IProjectable<TEntity, TResult>
    {
        return _dbSet
            .AsNoTracking()
            .Where(predicate)
            .Select(TResult.Projection)
            .ToArrayAsync();       
    }

    public Task<TResult[]> QueryReadOnlyAsync<TResult>(
        Func<IQueryable<TEntity>, IQueryable<TEntity>> queryBuilder)
        where TResult : class, IProjectable<TEntity, TResult>
    {
        var query = _dbSet.AsNoTracking();   
        query = queryBuilder(query);
        return query.Select(TResult.Projection).ToArrayAsync();
    }

    public Task<bool> ExistsAsync(int id) => _dbSet.AnyAsync(e => e.Id == id);
    public Task<bool> ExistsAsync(Expression<Func<TEntity, bool>> predicate) => _dbSet.AnyAsync(predicate);
    public Task<int> CountAsync() => _dbSet.CountAsync();
    public Task<int> CountAsync(Expression<Func<TEntity, bool>> predicate) => _dbSet.CountAsync(predicate);
    public ValueTask<TEntity?> GetWithTrackingAsync(int id) => _dbSet.FindAsync(id);
    public Task<List<TEntity>> GetListWithTrackingAsync(IReadOnlyCollection<int> ids) => _dbSet.Where(e => ids.Contains(e.Id)).ToListAsync();
    public Task<List<TEntity>> GetListWithTrackingAsync(Expression<Func<TEntity, bool>> predicate) => _dbSet.Where(predicate).ToListAsync();
    public TEntity Add(TEntity entity)
    {
        _dbSet.Add(entity);
        return entity;
    }
    public IReadOnlyCollection<TEntity> AddMultiple(IReadOnlyCollection<TEntity> entities)
    {
        _dbSet.AddRange(entities);
        return entities;
    }
    public void Remove(TEntity entity) => _dbSet.Remove(entity);
    public void RemoveMultiple(IReadOnlyCollection<TEntity> entities) => _dbSet.RemoveRange(entities);
    public Task SaveChangesAsync() => context.SaveChangesAsync();
    public void CancelChanges() => context.ChangeTracker.Clear();
}
