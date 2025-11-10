import React, { useState, useEffect } from 'react';
import type { DNSRecord, CreateDNSRecordRequest } from '../../types';
import Input from '../common/Input';
import Button from '../common/Button';

interface DNSRecordFormProps {
  record?: DNSRecord | null;
  onSubmit: (data: CreateDNSRecordRequest) => void;
  onCancel: () => void;
  isLoading: boolean;
  serverError?: string | null;
}

const DNSRecordForm: React.FC<DNSRecordFormProps> = ({ record, onSubmit, onCancel, isLoading, serverError }) => {
  const [formData, setFormData] = useState<CreateDNSRecordRequest>({
    DomainName: '',
    Type: 'A',
    Value: '',
  });
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  useEffect(() => {
    if (record) {
      setFormData({
        DomainName: record.DomainName,
        Type: record.Type,
        Value: record.Value,
      });
    } else {
      setFormData({ DomainName: '', Type: 'A', Value: '' });
    }
  }, [record]);

  const validate = (): boolean => {
    const newErrors: { [key: string]: string } = {};
    if (!formData.DomainName) newErrors.DomainName = 'Domain Name is required.';
    if (!formData.Value) newErrors.Value = 'Value is required.';
    // Basic validation, more complex regex can be added
    if (formData.Type === 'A' && !/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(formData.Value)) {
      newErrors.Value = 'Must be a valid IPv4 address for A record.';
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validate()) {
      onSubmit(formData);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  return (
    <form onSubmit={handleSubmit}>
      {serverError && <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">{serverError}</div>}
      <div className="mb-4">
        <Input
          label="Domain Name"
          name="DomainName"
          value={formData.DomainName}
          onChange={handleChange}
          error={errors.DomainName}
          placeholder="service.internal.local"
        />
      </div>
      <div className="mb-4">
        <label htmlFor="Type" className="block text-gray-700 text-sm font-bold mb-2">Record Type</label>
        <select
          id="Type"
          name="Type"
          value={formData.Type}
          onChange={handleChange}
          className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
        >
          <option value="A">A</option>
          <option value="CNAME">CNAME</option>
        </select>
      </div>
      <div className="mb-6">
        <Input
          label="Value"
          name="Value"
          value={formData.Value}
          onChange={handleChange}
          error={errors.Value}
          placeholder={formData.Type === 'A' ? '192.168.1.10' : 'target.internal.local'}
        />
      </div>
      <div className="flex items-center justify-end space-x-2">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={isLoading}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Save'}
        </Button>
      </div>
    </form>
  );
};

export default DNSRecordForm;

