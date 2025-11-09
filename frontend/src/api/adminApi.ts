import { axiosPrivate } from './axios';
import { User } from '../types';

export const getUsers = async (): Promise<User[]> => {
    const response = await axiosPrivate.get<User[]>('/admin/users');
    return response.data;
};

export const updateUserStatus = async (id: number, isEnabled: boolean): Promise<User> => {
    const response = await axiosPrivate.put<User>(`/admin/users/${id}/status`, { isEnabled });
    return response.data;
};

