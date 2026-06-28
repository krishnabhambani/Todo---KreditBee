import React, { useState, useContext } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import Toast from '../../components/Toast/Toast';
import { CheckSquare, LogIn, AlertCircle, Eye, EyeOff } from 'lucide-react';
import styles from './LoginPage.module.css';

const LoginPage = () => {
  const { login } = useContext(AuthContext);
  const navigate = useNavigate();

  const [formData, setFormData] = useState({ email: '', password: '' });
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const [toast, setToast] = useState(null);
  const [credentialError, setCredentialError] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  const validate = () => {
    const tempErrors = {};
    if (!formData.email) {
      tempErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      tempErrors.email = 'Invalid email address';
    }
    if (!formData.password) {
      tempErrors.password = 'Password is required';
    }
    setErrors(tempErrors);
    return Object.keys(tempErrors).length === 0;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
    if (errors[name]) {
      setErrors({ ...errors, [name]: '' });
    }
    // Clear credential error when user starts typing
    if (credentialError) {
      setCredentialError('');
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;

    setLoading(true);
    setCredentialError('');
    const result = await login(formData.email, formData.password);
    setLoading(false);

    if (result.success) {
      setToast({ type: 'success', message: 'Logged in successfully!' });
      setTimeout(() => navigate('/'), 1000);
    } else {
      // Set credential error for invalid login attempts
      if (result.message.toLowerCase().includes('invalid')) {
        setCredentialError(result.message);
      }
      setToast({ type: 'error', message: result.message });
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <div className={styles.header}>
          <div className={styles.logo}>
            <CheckSquare size={32} />
            <span>TodoApp</span>
          </div>
          <h1>Welcome Back</h1>
          <p>Please enter your credentials to login</p>
        </div>

        <form onSubmit={handleSubmit} className={styles.form}>
          {credentialError && (
            <div className={styles.credentialErrorAlert}>
              <AlertCircle size={16} style={{ marginRight: '8px' }} />
              <span>{credentialError}</span>
            </div>
          )}
          <div className={styles.inputGroup}>
            <label>Email Address</label>
            <input
              type="email"
              name="email"
              placeholder="e.g. user@example.com"
              value={formData.email}
              onChange={handleChange}
              className={`${errors.email ? styles.inputError : ''} ${credentialError ? styles.inputCredentialError : ''}`}
              disabled={loading}
            />
            {errors.email && <span className={styles.errorText}>{errors.email}</span>}
          </div>

          <div className={styles.inputGroup}>
            <label>Password</label>
            <div className={styles.passwordInputWrapper}>
              <input
                type={showPassword ? "text" : "password"}
                name="password"
                placeholder="••••••••"
                value={formData.password}
                onChange={handleChange}
                className={`${errors.password ? styles.inputError : ''} ${credentialError ? styles.inputCredentialError : ''}`}
                disabled={loading}
              />
              <button 
                type="button"
                className={styles.passwordToggleBtn}
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            {errors.password && <span className={styles.errorText}>{errors.password}</span>}
          </div>

          <button type="submit" className={styles.submitBtn} disabled={loading}>
            {loading ? (
              <span className="spinner" style={{ width: '20px', height: '20px' }}></span>
            ) : (
              <>
                <LogIn size={18} />
                <span>Login</span>
              </>
            )}
          </button>
        </form>

        <div className={styles.footer}>
          Don't have an account? <Link to="/register">Register here</Link>
        </div>
      </div>
      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  );
};

export default LoginPage;
