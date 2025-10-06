namespace API.DataAccess;

public interface IRepository<T>
{
    // For projections
    /// <summary>
    /// Returns the entities as an <see cref="IQueryable{T}"/> for advanced queries and projections.
    /// Uses AsNoTracking() for read-only scenarios.
    /// </summary>
    IQueryable<T> AsQueryable();
    // Create
    Task<T> CreateAsync(T entity);
    // Update
    Task UpdateAsync(T entity);
    // Delete
    Task DeleteAsync(T entity);
}
