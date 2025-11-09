import React, { useState, useEffect, useCallback } from 'react';
import { DNSRecord, CreateDNSRecordRequest } from '../types';
import * as dnsRecordApi from '../api/dnsRecordApi';
import Table from '../components/common/Table';
import Button from '../components/common/Button';
import Modal from '../components/common/Modal';
import DNSRecordForm from '../components/dns/DNSRecordForm';
import DeleteConfirmation from '../components/dns/DeleteConfirmation';
import Input from '../components/common/Input';
import { useDebounce } from '../hooks/useDebounce';

const DashboardPage = () => {
  const [records, setRecords] = useState<DNSRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [totalCount, setTotalCount] = useState(0);
  const [searchTerm, setSearchTerm] = useState('');
  const debouncedSearchTerm = useDebounce(searchTerm, 500);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalMode, setModalMode] = useState<'create' | 'edit' | 'delete' | null>(null);
  const [currentRecord, setCurrentRecord] = useState<DNSRecord | null>(null);
  const [formSubmitting, setFormSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const fetchRecords = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const { records: fetchedRecords, totalCount: fetchedTotalCount } = await dnsRecordApi.getDnsRecords({
        page,
        pageSize,
        search: debouncedSearchTerm,
      });
      setRecords(fetchedRecords);
      setTotalCount(fetchedTotalCount);
    } catch (err) {
      setError('Failed to fetch DNS records.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, debouncedSearchTerm]);

  useEffect(() => {
    fetchRecords();
  }, [fetchRecords]);

  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  const openModal = (mode: 'create' | 'edit' | 'delete', record: DNSRecord | null = null) => {
    setModalMode(mode);
    setCurrentRecord(record);
    setFormError(null);
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setModalMode(null);
    setCurrentRecord(null);
  };

  const handleFormSubmit = async (data: CreateDNSRecordRequest) => {
    setFormSubmitting(true);
    setFormError(null);
    try {
      if (modalMode === 'edit' && currentRecord) {
        await dnsRecordApi.updateDnsRecord(currentRecord.ID, data);
      } else {
        await dnsRecordApi.createDnsRecord(data);
      }
      closeModal();
      fetchRecords();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || `Failed to ${modalMode === 'edit' ? 'update' : 'create'} record.`;
      setFormError(errorMessage);
      console.error(err);
    } finally {
      setFormSubmitting(false);
    }
  };

  const handleDeleteConfirm = async () => {
    if (!currentRecord) return;
    setFormSubmitting(true);
    setFormError(null);
    try {
      await dnsRecordApi.deleteDnsRecord(currentRecord.ID);
      closeModal();
      fetchRecords();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || 'Failed to delete record.';
      setFormError(errorMessage);
      console.error(err);
    } finally {
      setFormSubmitting(false);
    }
  };

  const columns = [
    { header: 'Domain Name', accessor: 'DomainName' },
    { header: 'Type', accessor: 'Type' },
    { header: 'Value', accessor: 'Value' },
    {
      header: 'Actions',
      accessor: 'actions',
      render: (record: DNSRecord) => (
        <div className="space-x-2">
          <Button variant="secondary" onClick={() => openModal('edit', record)}>Edit</Button>
          <Button variant="secondary" className="bg-red-500 text-white hover:bg-red-600" onClick={() => openModal('delete', record)}>Delete</Button>
        </div>
      ),
    },
  ];

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">DNS Records Dashboard</h1>
      
      <div className="flex justify-between items-center mb-4">
        <div className="w-1/3">
          <Input 
            placeholder="Search by domain name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <Button onClick={() => openModal('create')}>Create New Record</Button>
      </div>

      {loading && <p>Loading records...</p>}
      {error && <p className="text-red-500">{error}</p>}
      {!loading && !error && (
        <Table
          columns={columns}
          data={records}
          totalCount={totalCount}
          page={page}
          pageSize={pageSize}
          onPageChange={handlePageChange}
        />
      )}

      <Modal
        isOpen={isModalOpen}
        onClose={closeModal}
        title={
          modalMode === 'create' ? 'Create DNS Record' :
          modalMode === 'edit' ? 'Edit DNS Record' : 'Confirm Deletion'
        }
      >
        {modalMode === 'create' || modalMode === 'edit' ? (
          <DNSRecordForm
            record={currentRecord}
            onSubmit={handleFormSubmit}
            onCancel={closeModal}
            isLoading={formSubmitting}
            serverError={formError}
          />
        ) : modalMode === 'delete' && currentRecord ? (
          <DeleteConfirmation
            record={currentRecord}
            onConfirm={handleDeleteConfirm}
            onCancel={closeModal}
            isLoading={formSubmitting}
            serverError={formError}
          />
        ) : null}
      </Modal>
    </div>
  );
};

export default DashboardPage;

