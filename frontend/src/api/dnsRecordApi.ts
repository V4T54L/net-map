import { axiosPrivate } from './axios';
import { DNSRecord, CreateDNSRecordRequest, UpdateDNSRecordRequest } from '../types';

interface GetDnsRecordsResponse {
  records: DNSRecord[];
  totalCount: number;
}

export const getDnsRecords = async (
  params: { page: number; pageSize: number; search?: string }
): Promise<GetDnsRecordsResponse> => {
  const response = await axiosPrivate.get('/dns-records', { params });
  const totalCount = parseInt(response.headers['x-total-count'] || '0', 10);
  return { records: response.data, totalCount };
};

export const createDnsRecord = async (data: CreateDNSRecordRequest): Promise<DNSRecord> => {
  const response = await axiosPrivate.post('/dns-records', data);
  return response.data;
};

export const updateDnsRecord = async (id: number, data: UpdateDNSRecordRequest): Promise<DNSRecord> => {
  const response = await axiosPrivate.put(`/dns-records/${id}`, data);
  return response.data;
};

export const deleteDnsRecord = async (id: number): Promise<void> => {
  await axiosPrivate.delete(`/dns-records/${id}`);
};

