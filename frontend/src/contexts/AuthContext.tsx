import React, { createContext, useState, useEffect, ReactNode } from 'react';
import { jwtDecode } from 'jwt-decode';
import { AuthContextType, User, LoginRequest, RegisterRequest, AuthTokens } from '../types';
import * as authApi from '../api/authApi';

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [tokens, setTokens] = useState<AuthTokens | null>(() => {
    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');
    return accessToken && refreshToken ? { AccessToken: accessToken, RefreshToken: refreshToken } : null;
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (tokens?.AccessToken) {
      try {
        const decoded: User = jwtDecode(tokens.AccessToken);
        // Check if token is expired
        const isExpired = decoded.exp ? decoded.exp * 1000 < Date.now() : true;
        if (!isExpired) {
          setUser({
            ID: decoded.UserID, // Mapping from token claims
            Username: decoded.sub, // Subject is username
            Role: decoded.Role,
            IsEnabled: true, // Assuming token means enabled
          });
        } else {
          // Handle expired token, maybe try to refresh
          logout();
        }
      } catch (error) {
        console.error('Invalid token:', error);
        logout();
      }
    }
    setLoading(false);
  }, [tokens]);

  const login = async (username: string, password: string): Promise<void> => {
    const loginData: LoginRequest = { Username: username, Password: password };
    const response = await authApi.loginUser(loginData);
    localStorage.setItem('accessToken', response.AccessToken);
    localStorage.setItem('refreshToken', response.RefreshToken);
    setTokens(response);
  };

  const register = async (username: string, password: string): Promise<void> => {
    const registerData: RegisterRequest = { Username: username, Password: password };
    await authApi.registerUser(registerData);
  };

  const logout = () => {
    setUser(null);
    setTokens(null);
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  };

  const value = { user, tokens, login, logout, register, loading };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthContext;

