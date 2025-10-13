namespace API.DataAccess;

public interface IRepository<T> where T : class
{
    IQueryable<T> AsQueryable();
    // Read
    Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : IProjectable<T, TProjectable>;
    Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : IProjectable<T, TProjectable>;
    /// <summary>
    /// Gets the object without joins.
    /// </summary>
    /// <param name="id"></param>
    /// <returns></returns>
    Task<T?> GetAsync(int id);
    // Create
    Task<T> CreateAsync(T entity);
    // Update
    Task<T> UpdateAsync(T updatedEntity);
    // Delete
    Task DeleteAsync(T entity);
}
