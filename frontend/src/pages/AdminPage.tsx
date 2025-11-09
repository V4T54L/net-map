import React, { useState } from 'react';
import UserManagementTable from '../components/admin/UserManagementTable';
import AllDnsRecordsTable from '../components/admin/AllDnsRecordsTable';

const AdminPage: React.FC = () => {
    const [activeTab, setActiveTab] = useState<'users' | 'dns'>('users');

    return (
        <div className="container mx-auto p-4">
            <h1 className="text-2xl font-bold mb-4">Admin Dashboard</h1>

            <div className="border-b border-gray-200">
                <nav className="-mb-px flex space-x-8" aria-label="Tabs">
                    <button
                        onClick={() => setActiveTab('users')}
                        className={`${
                            activeTab === 'users'
                                ? 'border-indigo-500 text-indigo-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                        } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
                    >
                        User Management
                    </button>
                    <button
                        onClick={() => setActiveTab('dns')}
                        className={`${
                            activeTab === 'dns'
                                ? 'border-indigo-500 text-indigo-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                        } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm`}
                    >
                        All DNS Records
                    </button>
                </nav>
            </div>

            <div className="mt-8">
                {activeTab === 'users' && <UserManagementTable />}
                {activeTab === 'dns' && <AllDnsRecordsTable />}
            </div>
        </div>
    );
};

export default AdminPage;

