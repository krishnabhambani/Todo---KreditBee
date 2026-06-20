import axios from 'axios';

const API = axios.create({
  baseURL: import.meta.env.VITE_API_URL + '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request Interceptor: Automatically inject Authorization token if exists
API.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response Interceptor: Catch auth errors (401) to auto-logout user
API.interceptors.response.use(
  (response) => response.data, // return the response data directly
  (error) => {
    if (error.response && error.response.status === 401) {
      // Check if the 401 error came from the login request itself
      const isLoginRequest = error.config.url.includes('/auth/login');

      // ONLY redirect and clear localStorage if it's NOT a login request
      if (!isLoginRequest) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        // Force page refresh to redirect to login via ProtectedRoute
        window.location.href = '/login';
      }
    }
    
    const message = error.response?.data?.message || 'Something went wrong. Please try again.';
    return Promise.reject(new Error(message));
  }
);

export default API;