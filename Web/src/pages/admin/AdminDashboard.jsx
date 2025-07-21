import React from 'react';
import { Link } from 'react-router';

const AdminDashboard = () => {
    return (
        <div className="admin-dashboard">
            <h1>Admin Dashboard</h1>
            <section>
                <Link to="/admin/players">Player Management</Link>
            </section>
            <section>
                <Link to="/admin/roles">Role Management</Link>
            </section>
            <section>
                <Link to="/admin/abilities">Ability Management</Link>
            </section>
        </div>
    );
};

export default AdminDashboard;