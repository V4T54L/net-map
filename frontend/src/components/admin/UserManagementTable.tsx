import React, { useState, useEffect, useCallback } from 'react';
import { getUsers, updateUserStatus } from '../../api/adminApi';
import type { User } from '../../types';
import Table from '../common/Table';
import type { Column } from '../common/Table';

const UserManagementTable: React.FC = () => {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchUsers = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await getUsers();
            setUsers(data);
        } catch (err) {
            setError('Failed to fetch users.');
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchUsers();
    }, [fetchUsers]);

    const handleToggleStatus = async (user: User) => {
        try {
            await updateUserStatus(user.ID, !user.IsEnabled);
            // Refresh the list to show the updated status
            fetchUsers();
        } catch (err) {
            alert(`Failed to update status for user ${user.Username}.`);
            console.error(err);
        }
    };

    const columns: Column[] = [
        { header: 'Username', accessor: 'Username' },
        { header: 'Role', accessor: 'Role' },
        {
            header: 'Status',
            accessor: 'IsEnabled',
            render: (row: User) => (
                <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                    row.IsEnabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                    {row.IsEnabled ? 'Enabled' : 'Disabled'}
                </span>
            ),
        },
        {
            header: 'Actions',
            accessor: 'actions',
            render: (row: User) => (
                <button
                    onClick={() => handleToggleStatus(row)}
                    className="text-indigo-600 hover:text-indigo-900 disabled:opacity-50"
                >
                    {row.IsEnabled ? 'Disable' : 'Enable'}
                </button>
            ),
        },
    ];

    if (loading) return <p>Loading users...</p>;
    if (error) return <p className="text-red-500">{error}</p>;

    return (
        <div>
            <h2 className="text-xl font-semibold mb-4">Users</h2>
            <Table
                columns={columns}
                data={users}
                totalCount={users.length}
                page={1}
                pageSize={users.length} // No pagination for now
                onPageChange={() => {}}
            />
        </div>
    );
};

export default UserManagementTable;

