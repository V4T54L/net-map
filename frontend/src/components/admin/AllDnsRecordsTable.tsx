import React, { useState, useEffect, useCallback } from 'react';
import { getDnsRecords, createDnsRecord, updateDnsRecord, deleteDnsRecord } from '../../api/dnsRecordApi';
import { DNSRecord, CreateDNSRecordRequest } from '../../types';
import Table, { Column } from '../common/Table';
import Modal from '../common/Modal';
import DNSRecordForm from '../dns/DNSRecordForm';
import DeleteConfirmation from '../dns/DeleteConfirmation';
import useDebounce from '../../hooks/useDebounce';

const AllDnsRecordsTable: React.FC = () => {
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
            // Admins will get all records from this endpoint
            const { records: fetchedRecords, totalCount: fetchedTotalCount } = await getDnsRecords({
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
            if (modalMode === 'create') {
                // Admin creates a record for themselves
                await createDnsRecord(data);
            } else if (modalMode === 'edit' && currentRecord) {
                await updateDnsRecord(currentRecord.ID, data);
            }
            closeModal();
            fetchRecords();
        } catch (err: any) {
            const errorMessage = err.response?.data?.message || `Failed to ${modalMode} record.`;
            setFormError(errorMessage);
        } finally {
            setFormSubmitting(false);
        }
    };

    const handleDeleteConfirm = async () => {
        if (!currentRecord) return;
        setFormSubmitting(true);
        setFormError(null);
        try {
            await deleteDnsRecord(currentRecord.ID);
            closeModal();
            fetchRecords();
        } catch (err: any) {
            const errorMessage = err.response?.data?.message || 'Failed to delete record.';
            setFormError(errorMessage);
        } finally {
            setFormSubmitting(false);
        }
    };

    const columns: Column[] = [
        { header: 'Domain Name', accessor: 'DomainName' },
        { header: 'Type', accessor: 'Type' },
        { header: 'Value', accessor: 'Value' },
        { header: 'Owner', accessor: 'Username' }, // New column for admin view
        {
            header: 'Actions',
            accessor: 'actions',
            render: (row: DNSRecord) => (
                <div className="space-x-2">
                    <button onClick={() => openModal('edit', row)} className="text-indigo-600 hover:text-indigo-900">
                        Edit
                    </button>
                    <button onClick={() => openModal('delete', row)} className="text-red-600 hover:text-red-900">
                        Delete
                    </button>
                </div>
            ),
        },
    ];

    if (loading) return <p>Loading records...</p>;
    if (error) return <p className="text-red-500">{error}</p>;

    return (
        <div>
            <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold">All DNS Records</h2>
                <button onClick={() => openModal('create')} className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700">
                    Create New Record
                </button>
            </div>
            <div className="mb-4">
                <input
                    type="text"
                    placeholder="Search domains..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="w-full p-2 border border-gray-300 rounded-md"
                />
            </div>
            <Table
                columns={columns}
                data={records}
                totalCount={totalCount}
                page={page}
                pageSize={pageSize}
                onPageChange={handlePageChange}
            />
            <Modal isOpen={isModalOpen} onClose={closeModal} title={
                modalMode === 'create' ? 'Create DNS Record' :
                modalMode === 'edit' ? 'Edit DNS Record' : 'Delete DNS Record'
            }>
                {modalMode === 'create' && <DNSRecordForm onCancel={closeModal} onSubmit={handleFormSubmit} isLoading={formSubmitting} serverError={formError || undefined} />}
                {modalMode === 'edit' && currentRecord && <DNSRecordForm record={currentRecord} onCancel={closeModal} onSubmit={handleFormSubmit} isLoading={formSubmitting} serverError={formError || undefined} />}
                {modalMode === 'delete' && currentRecord && <DeleteConfirmation record={currentRecord} onCancel={closeModal} onConfirm={handleDeleteConfirm} isLoading={formSubmitting} serverError={formError || undefined} />}
            </Modal>
        </div>
    );
};

export default AllDnsRecordsTable;

