import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import DashboardPage from './DashboardPage';
import * as dnsRecordApi from '../api/dnsRecordApi';
import { DNSRecord } from '../types';

jest.mock('../api/dnsRecordApi');
const mockedDnsRecordApi = dnsRecordApi as jest.Mocked<typeof dnsRecordApi>;

const mockRecords: DNSRecord[] = [
  { ID: 1, UserID: 1, DomainName: 'test1.local', Type: 'A', Value: '1.1.1.1', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString() },
  { ID: 2, UserID: 1, DomainName: 'test2.local', Type: 'CNAME', Value: 'test1.local', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString() },
];

const renderComponent = () => {
  return render(
    <BrowserRouter>
      <DashboardPage />
    </BrowserRouter>
  );
};

describe('DashboardPage', () => {
  beforeEach(() => {
    mockedDnsRecordApi.getDnsRecords.mockResolvedValue({ records: mockRecords, totalCount: mockRecords.length });
    mockedDnsRecordApi.createDnsRecord.mockResolvedValue(mockRecords[0]);
    mockedDnsRecordApi.updateDnsRecord.mockResolvedValue(mockRecords[0]);
    mockedDnsRecordApi.deleteDnsRecord.mockResolvedValue();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('renders dashboard and fetches records', async () => {
    renderComponent();
    expect(screen.getByText('DNS Records Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Loading records...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('test1.local')).toBeInTheDocument();
      expect(screen.getByText('test2.local')).toBeInTheDocument();
    });
  });

  test('opens create modal and submits new record', async () => {
    renderComponent();
    await waitFor(() => expect(screen.getByText('test1.local')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Create New Record'));
    expect(screen.getByText('Create DNS Record')).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('Domain Name'), { target: { value: 'new.local' } });
    fireEvent.change(screen.getByLabelText('Value'), { target: { value: '2.2.2.2' } });
    
    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => {
      expect(mockedDnsRecordApi.createDnsRecord).toHaveBeenCalledWith({
        DomainName: 'new.local',
        Type: 'A',
        Value: '2.2.2.2',
      });
      expect(screen.queryByText('Create DNS Record')).not.toBeInTheDocument();
    });
  });

  test('opens edit modal and submits updated record', async () => {
    renderComponent();
    await waitFor(() => expect(screen.getByText('test1.local')).toBeInTheDocument());

    const editButtons = screen.getAllByText('Edit');
    fireEvent.click(editButtons[0]);

    expect(screen.getByText('Edit DNS Record')).toBeInTheDocument();
    const domainInput = screen.getByLabelText('Domain Name') as HTMLInputElement;
    expect(domainInput.value).toBe('test1.local');

    fireEvent.change(domainInput, { target: { value: 'updated.local' } });
    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => {
      expect(mockedDnsRecordApi.updateDnsRecord).toHaveBeenCalledWith(1, {
        DomainName: 'updated.local',
        Type: 'A',
        Value: '1.1.1.1',
      });
    });
  });

  test('opens delete modal and confirms deletion', async () => {
    renderComponent();
    await waitFor(() => expect(screen.getByText('test1.local')).toBeInTheDocument());

    const deleteButtons = screen.getAllByText('Delete');
    fireEvent.click(deleteButtons[0]);

    expect(screen.getByText('Confirm Deletion')).toBeInTheDocument();
    expect(screen.getByText(/Are you sure you want to delete the record for/)).toBeInTheDocument();
    
    const confirmDeleteButton = screen.getAllByText('Delete').find(btn => btn.closest('.bg-red-600'));
    expect(confirmDeleteButton).toBeInTheDocument();
    fireEvent.click(confirmDeleteButton!);

    await waitFor(() => {
      expect(mockedDnsRecordApi.deleteDnsRecord).toHaveBeenCalledWith(1);
    });
  });
});

