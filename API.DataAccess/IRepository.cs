using API.Domain;

namespace API.DataAccess;

public interface IRepository<T> where T : class
{
    // Read
    Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : class, IProjectable<T, TProjectable>;
    Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : class, IProjectable<T, TProjectable>;
    /// <summary>
    /// Gets the object without joins.
    /// </summary>
    /// <param name="id"></param>
    /// <returns></returns>
    Task<T?> GetAsync(int id);
    Task<Result<List<T>>> GetMultipleAsync(List<int> ids);
    // Create
    Task<T> CreateAsync(T entity);
    // Update
    Task<T> UpdateAsync(T updatedEntity);
    // Delete
    Task DeleteAsync(T entity);
}
