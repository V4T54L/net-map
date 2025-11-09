import axios from 'axios';
import { jwtDecode } from 'jwt-decode';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export const axiosPublic = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

export const axiosPrivate = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

axiosPrivate.interceptors.request.use(
  (config) => {
    const accessToken = localStorage.getItem('accessToken');
    if (accessToken) {
      config.headers['Authorization'] = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

axiosPrivate.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    if (error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          // This endpoint does not exist yet, but is planned for JWT refresh
          // For now, we'll simulate a failure and logout
          // const { data } = await axiosPublic.post('/auth/refresh', { refreshToken });
          // localStorage.setItem('accessToken', data.accessToken);
          // originalRequest.headers['Authorization'] = `Bearer ${data.accessToken}`;
          // return axiosPrivate(originalRequest);
          
          // Placeholder for refresh logic: for now, just log out
          console.error("Access token expired, refresh mechanism not implemented. Logging out.");
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          window.location.href = '/login';
          return Promise.reject(error);

        } catch (refreshError) {
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          window.location.href = '/login';
          return Promise.reject(refreshError);
        }
      }
    }
    return Promise.reject(error);
  }
);

