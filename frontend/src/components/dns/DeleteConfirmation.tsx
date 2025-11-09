import React from 'react';
import { DNSRecord } from '../../types';
import Button from '../common/Button';

interface DeleteConfirmationProps {
  record: DNSRecord;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading: boolean;
  serverError?: string | null;
}

const DeleteConfirmation: React.FC<DeleteConfirmationProps> = ({ record, onConfirm, onCancel, isLoading, serverError }) => {
  return (
    <div>
      {serverError && <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">{serverError}</div>}
      <p className="text-gray-700 mb-4">
        Are you sure you want to delete the record for <strong className="font-semibold">{record.DomainName}</strong>?
      </p>
      <div className="flex items-center justify-end space-x-2">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={isLoading}>
          Cancel
        </Button>
        <Button type="button" variant="primary" onClick={onConfirm} disabled={isLoading} className="bg-red-600 hover:bg-red-700">
          {isLoading ? 'Deleting...' : 'Delete'}
        </Button>
      </div>
    </div>
  );
};

export default DeleteConfirmation;

