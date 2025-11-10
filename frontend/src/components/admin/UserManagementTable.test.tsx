import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import UserManagementTable from './UserManagementTable';
import * as adminApi from '../../api/adminApi';
import type{ User } from '../../types';

jest.mock('../../api/adminApi');
const mockedAdminApi = adminApi as jest.Mocked<typeof adminApi>;

const mockUsers: User[] = [
    { ID: 1, Username: 'admin', Role: 'admin', IsEnabled: true, CreatedAt: '', UpdatedAt: '' },
    { ID: 2, Username: 'user1', Role: 'user', IsEnabled: true, CreatedAt: '', UpdatedAt: '' },
    { ID: 3, Username: 'user2', Role: 'user', IsEnabled: false, CreatedAt: '', UpdatedAt: '' },
];

describe('UserManagementTable', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedAdminApi.getUsers.mockResolvedValue(mockUsers);
        mockedAdminApi.updateUserStatus.mockImplementation(async (id, isEnabled) => {
            const user = mockUsers.find(u => u.ID === id);
            if (!user) throw new Error("User not found");
            return { ...user, IsEnabled: isEnabled };
        });
    });

    const renderComponent = () => {
        render(<UserManagementTable />);
    };

    test('fetches and displays users on mount', async () => {
        renderComponent();
        expect(screen.getByText('Loading users...')).toBeInTheDocument();
        
        await waitFor(() => {
            expect(screen.getByText('admin')).toBeInTheDocument();
            expect(screen.getByText('user1')).toBeInTheDocument();
            expect(screen.getByText('user2')).toBeInTheDocument();
        });

        expect(mockedAdminApi.getUsers).toHaveBeenCalledTimes(1);
    });

    test('displays user status correctly', async () => {
        renderComponent();
        await waitFor(() => {
            // user1 is enabled
            const user1Row = screen.getByText('user1').closest('tr');
            expect(user1Row).toHaveTextContent('Enabled');
            // user2 is disabled
            const user2Row = screen.getByText('user2').closest('tr');
            expect(user2Row).toHaveTextContent('Disabled');
        });
    });

    test('calls updateUserStatus and refetches users when disable button is clicked', async () => {
        renderComponent();
        await waitFor(() => expect(screen.getByText('user1')).toBeInTheDocument());

        const disableButtons = screen.getAllByRole('button', { name: 'Disable' });
        // Assuming user1 is the first 'Disable' button
        fireEvent.click(disableButtons[0]);

        await waitFor(() => {
            expect(mockedAdminApi.updateUserStatus).toHaveBeenCalledWith(2, false);
        });

        // It should refetch users after update
        await waitFor(() => {
            expect(mockedAdminApi.getUsers).toHaveBeenCalledTimes(2);
        });
    });

    test('calls updateUserStatus when enable button is clicked', async () => {
        renderComponent();
        await waitFor(() => expect(screen.getByText('user2')).toBeInTheDocument());

        const enableButton = screen.getByRole('button', { name: 'Enable' });
        fireEvent.click(enableButton);

        await waitFor(() => {
            expect(mockedAdminApi.updateUserStatus).toHaveBeenCalledWith(3, true);
        });
    });

    test('displays an error message if fetching users fails', async () => {
        mockedAdminApi.getUsers.mockRejectedValue(new Error('API Error'));
        renderComponent();

        await waitFor(() => {
            expect(screen.getByText('Failed to fetch users.')).toBeInTheDocument();
        });
    });
});

