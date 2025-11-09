import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import AdminPage from './AdminPage';

// Mock child components
jest.mock('../components/admin/UserManagementTable', () => () => <div>User Management Table</div>);
jest.mock('../components/admin/AllDnsRecordsTable', () => () => <div>All DNS Records Table</div>);

describe('AdminPage', () => {
    const renderComponent = () => {
        render(
            <BrowserRouter>
                <AdminPage />
            </BrowserRouter>
        );
    };

    test('renders admin dashboard title and tabs', () => {
        renderComponent();
        expect(screen.getByText('Admin Dashboard')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'User Management' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'All DNS Records' })).toBeInTheDocument();
    });

    test('shows User Management table by default', () => {
        renderComponent();
        expect(screen.getByText('User Management Table')).toBeInTheDocument();
        expect(screen.queryByText('All DNS Records Table')).not.toBeInTheDocument();
    });

    test('switches to All DNS Records tab on click', () => {
        renderComponent();
        const dnsTab = screen.getByRole('button', { name: 'All DNS Records' });
        fireEvent.click(dnsTab);

        expect(screen.getByText('All DNS Records Table')).toBeInTheDocument();
        expect(screen.queryByText('User Management Table')).not.toBeInTheDocument();
    });

    test('switches back to User Management tab on click', () => {
        renderComponent();
        const dnsTab = screen.getByRole('button', { name: 'All DNS Records' });
        fireEvent.click(dnsTab);
        const userTab = screen.getByRole('button', { name: 'User Management' });
        fireEvent.click(userTab);

        expect(screen.getByText('User Management Table')).toBeInTheDocument();
        expect(screen.queryByText('All DNS Records Table')).not.toBeInTheDocument();
    });
});

