import { render, screen, waitFor } from '@testing-library/react';
import AllDnsRecordsTable from './AllDnsRecordsTable';
import * as dnsRecordApi from '../../api/dnsRecordApi';
import type { AuthTokens, DNSRecord, User } from '../../types';
import AuthContext from '../../contexts/AuthContext';

jest.mock('../../api/dnsRecordApi');
const mockedDnsRecordApi = dnsRecordApi as jest.Mocked<typeof dnsRecordApi>;

const mockRecords: DNSRecord[] = [
    { ID: 1, UserID: 1, Username: 'admin', DomainName: 'service1.internal', Type: 'A', Value: '10.0.0.1', CreatedAt: '', UpdatedAt: '' },
    { ID: 2, UserID: 2, Username: 'user1', DomainName: 'service2.internal', Type: 'CNAME', Value: 'service1.internal', CreatedAt: '', UpdatedAt: '' },
];

const mockAuthContext = {
    user: { ID: 1, Username: 'admin', Role: 'admin' as "admin" | "user", IsEnabled: true, CreatedAt: "", UpdatedAt: "" },
    tokens: { accessToken: 'fake-token', refreshToken: 'fake-token' },
    loading: false,
    login: jest.fn(),
    logout: jest.fn(),
    register: jest.fn(),
};

describe('AllDnsRecordsTable', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedDnsRecordApi.getDnsRecords.mockResolvedValue({ records: mockRecords, totalCount: mockRecords.length });
    });

    const renderComponent = () => {
        render(
            <AuthContext.Provider value={mockAuthContext}>
                <AllDnsRecordsTable />
            </AuthContext.Provider>
        );
    };

    test('fetches and displays all DNS records, including owner username', async () => {
        renderComponent();
        expect(screen.getByText('Loading records...')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.getByText('service1.internal')).toBeInTheDocument();
            expect(screen.getByText('service2.internal')).toBeInTheDocument();
            // Check for owner username column
            expect(screen.getByText('admin')).toBeInTheDocument();
            expect(screen.getByText('user1')).toBeInTheDocument();
        });

        expect(mockedDnsRecordApi.getDnsRecords).toHaveBeenCalledTimes(1);
    });

    test('displays an error message if fetching records fails', async () => {
        mockedDnsRecordApi.getDnsRecords.mockRejectedValue(new Error('API Error'));
        renderComponent();

        await waitFor(() => {
            expect(screen.getByText('Failed to fetch DNS records.')).toBeInTheDocument();
        });
    });
});

