import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import DNSRecordForm from './DNSRecordForm';

describe('DNSRecordForm', () => {
  const mockOnSubmit = jest.fn();
  const mockOnCancel = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders create form correctly', () => {
    render(<DNSRecordForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} isLoading={false} />);
    expect(screen.getByLabelText('Domain Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Record Type')).toBeInTheDocument();
    expect(screen.getByLabelText('Value')).toBeInTheDocument();
    expect(screen.getByText('Save')).toBeInTheDocument();
  });

  test('renders edit form with initial data', () => {
    const mockRecord = {
      ID: 1, UserID: 1, DomainName: 'edit.local', Type: 'CNAME' as 'A' | 'CNAME', Value: 'target.local', CreatedAt: '', UpdatedAt: ''
    };
    render(<DNSRecordForm record={mockRecord} onSubmit={mockOnSubmit} onCancel={mockOnCancel} isLoading={false} />);
    
    expect(screen.getByLabelText<HTMLInputElement>('Domain Name').value).toBe('edit.local');
    expect(screen.getByLabelText<HTMLSelectElement>('Record Type').value).toBe('CNAME');
    expect(screen.getByLabelText<HTMLInputElement>('Value').value).toBe('target.local');
  });

  test('shows validation errors for empty fields', () => {
    render(<DNSRecordForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} isLoading={false} />);
    fireEvent.click(screen.getByText('Save'));
    
    expect(screen.getByText('Domain Name is required.')).toBeInTheDocument();
    expect(screen.getByText('Value is required.')).toBeInTheDocument();
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  test('shows validation error for invalid IPv4 for A record', () => {
    render(<DNSRecordForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} isLoading={false} />);
    
    fireEvent.change(screen.getByLabelText('Domain Name'), { target: { value: 'test.local' } });
    fireEvent.change(screen.getByLabelText('Record Type'), { target: { value: 'A' } });
    fireEvent.change(screen.getByLabelText('Value'), { target: { value: 'invalid-ip' } });
    
    fireEvent.click(screen.getByText('Save'));
    
    expect(screen.getByText('Must be a valid IPv4 address for A record.')).toBeInTheDocument();
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  test('calls onSubmit with form data when valid', () => {
    render(<DNSRecordForm onSubmit={mockOnSubmit} onCancel={mockOnCancel} isLoading={false} />);
    
    fireEvent.change(screen.getByLabelText('Domain Name'), { target: { value: 'valid.local' } });
    fireEvent.change(screen.getByLabelText('Record Type'), { target: { value: 'A' } });
    fireEvent.change(screen.getByLabelText('Value'), { target: { value: '1.2.3.4' } });
    
    fireEvent.click(screen.getByText('Save'));
    
    expect(mockOnSubmit).toHaveBeenCalledWith({
      DomainName: 'valid.local',
      Type: 'A',
      Value: '1.2.3.4',
    });
  });
});

