import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import Button from '../common/Button';

const Navbar: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="bg-white shadow-md">
      <div className="container mx-auto px-4 py-2 flex justify-between items-center">
        <Link to="/" className="text-xl font-bold text-blue-600">
          Internal DNS
        </Link>
        <div>
          {user ? (
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">Welcome, {user.Username}</span>
              {user.Role === 'admin' && (
                <Link to="/admin" className="text-blue-500 hover:underline">
                  Admin
                </Link>
              )}
              <Button onClick={handleLogout} className="w-auto px-4 py-2">
                Logout
              </Button>
            </div>
          ) : (
            <div className="space-x-2">
              <Link to="/login">
                <Button className="w-auto px-4 py-2">Login</Button>
              </Link>
              <Link to="/register">
                <Button variant="secondary" className="w-auto px-4 py-2">
                  Register
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;

