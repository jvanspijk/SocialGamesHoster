namespace API.DataAccess;

public interface IRepository<T> where T : class
{
    IQueryable<T> AsQueryable();
    // Read
    Task<TProjection?> GetAsync<TProjection>(int id) where TProjection : IProjectable<T, TProjection>;
    Task<List<TProjection>> GetAllAsync<TProjection>() where TProjection : IProjectable<T, TProjection>;
    // Create
    Task<T> CreateAsync(T entity);
    // Update
    Task UpdateAsync(T entity);
    // Delete
    Task DeleteAsync(T entity);
}
