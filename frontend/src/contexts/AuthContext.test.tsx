import React from 'react';
import { render, screen, act } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import * as authApi from '../api/authApi';
import { User } from '../types';

jest.mock('../api/authApi');
const mockedAuthApi = authApi as jest.Mocked<typeof authApi>;

const mockUser: User = { ID: 1, Username: 'testuser', Role: 'user', IsEnabled: true };

// Mock jwt-decode
jest.mock('jwt-decode', () => ({
  jwtDecode: () => ({
    UserID: 1,
    Username: 'testuser',
    Role: 'user',
    IsEnabled: true,
    exp: Date.now() / 1000 + 3600,
  }),
}));

const TestComponent = () => {
  const { user, login, logout, register } = useAuth();
  return (
    <div>
      <div data-testid="user">{user ? user.Username : 'null'}</div>
      <button onClick={() => login('testuser', 'password')}>Login</button>
      <button onClick={() => logout()}>Logout</button>
      <button onClick={() => register('newuser', 'password')}>Register</button>
    </div>
  );
};

describe('AuthContext', () => {
  beforeEach(() => {
    localStorage.clear();
    jest.clearAllMocks();
  });

  test('login successfully updates user state and localStorage', async () => {
    mockedAuthApi.loginUser.mockResolvedValue({
      AccessToken: 'fake-access-token',
      RefreshToken: 'fake-refresh-token',
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await act(async () => {
      screen.getByText('Login').click();
    });

    expect(screen.getByTestId('user')).toHaveTextContent('testuser');
    expect(localStorage.getItem('accessToken')).toBe('fake-access-token');
    expect(localStorage.getItem('refreshToken')).toBe('fake-refresh-token');
  });

  test('logout clears user state and localStorage', async () => {
    // First, log in
    mockedAuthApi.loginUser.mockResolvedValue({
      AccessToken: 'fake-access-token',
      RefreshToken: 'fake-refresh-token',
    });
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );
    await act(async () => {
      screen.getByText('Login').click();
    });
    expect(screen.getByTestId('user')).toHaveTextContent('testuser');

    // Then, log out
    await act(async () => {
      screen.getByText('Logout').click();
    });

    expect(screen.getByTestId('user')).toHaveTextContent('null');
    expect(localStorage.getItem('accessToken')).toBeNull();
    expect(localStorage.getItem('refreshToken')).toBeNull();
  });

  test('register calls the register API', async () => {
    mockedAuthApi.registerUser.mockResolvedValue();
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await act(async () => {
      screen.getByText('Register').click();
    });

    expect(mockedAuthApi.registerUser).toHaveBeenCalledWith({
      Username: 'newuser',
      Password: 'password',
    });
  });
});

