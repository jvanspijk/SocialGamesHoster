namespace API.DataAccess;

public interface IRepository<T> where T : class
{
    IQueryable<T> AsQueryable();
    // Read
    Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : IProjectable<T, TProjectable>;
    Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : IProjectable<T, TProjectable>;
    // Create
    Task<T> CreateAsync(T entity);
    // Update
    Task UpdateAsync(T entity);
    // Delete
    Task DeleteAsync(T entity);
}
