import React, { createContext, useState, useEffect, type ReactNode, useContext } from 'react';
import type { AuthContextType, User, LoginRequest, RegisterRequest, AuthTokens } from '../types';
import * as authApi from '../api/authApi';
import { jwtDecode } from 'jwt-decode';

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [tokens, setTokens] = useState<AuthTokens | null>(() => {
    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');
    return accessToken && refreshToken ? { accessToken: accessToken, refreshToken: refreshToken } : null;
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (tokens?.accessToken) {
      try {
        const decoded: {
          user_id: number,
          role: string,
          sub: string,
          exp: number,
          iat: number,
        } = jwtDecode(tokens.accessToken);
        // Check if token is expired
        const isExpired = decoded.exp ? decoded.exp * 1000 < Date.now() : true;
        // const isExpired = false
        if (!isExpired) {
          setUser({
            ID: decoded.user_id, // Mapping from token claims
            Username: decoded.sub, // Subject is username
            Role: decoded.role as "user" | "admin",
            IsEnabled: true, // Assuming token means enabled
            CreatedAt: "",
            UpdatedAt: ""
          });
        } else {
          // Handle expired token, maybe try to refresh
          logout();
        }
      } catch (error) {
        console.error('Invalid token:', error);
        // logout();
      }
    }
    setLoading(false);
  }, [tokens]);

  const login = async (username: string, password: string): Promise<void> => {
    const loginData: LoginRequest = { Username: username, Password: password };
    const response = await authApi.loginUser(loginData);
    localStorage.setItem('accessToken', response.accessToken);
    localStorage.setItem('refreshToken', response.refreshToken);
    setTokens(response);
  };

  const register = async (username: string, password: string): Promise<void> => {
    const registerData: RegisterRequest = { Username: username, Password: password };
    await authApi.registerUser(registerData);
  };

  const logout = () => {
    // setUser(null);
    // setTokens(null);
    // localStorage.removeItem('accessToken');
    // localStorage.removeItem('refreshToken');
  };

  const value = { user, tokens, login, logout, register, loading };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};


export default AuthContext;

