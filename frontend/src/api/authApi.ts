import { LoginRequest, RegisterRequest, AuthResponse } from '../types';
import { axiosPublic } from './axios';

export const registerUser = async (data: RegisterRequest): Promise<void> => {
  await axiosPublic.post('/auth/register', data);
};

export const loginUser = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await axiosPublic.post<AuthResponse>('/auth/login', data);
  return response.data;
};

