import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import API from '../../services/api';
import Toast from '../../components/Toast/Toast';
import styles from './UpdatePasswordPage.module.css';

const UpdatePasswordPage = () => {
  const navigate = useNavigate();
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [toast, setToast] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      setToast({ message: 'New password and confirm do not match', type: 'error' });
      return;
    }

    setLoading(true);
    try {
      await API.patch('/auth/password', { current_password: currentPassword, new_password: newPassword });
      setToast({ message: 'Password updated successfully', type: 'success' });
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      // optional: navigate back to profile after short delay
      setTimeout(() => navigate('/profile'), 1000);
    } catch (err) {
      setToast({ message: err.message || 'Failed to update password', type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <h2 className={styles.title}>Update Password</h2>
      <form className={styles.form} onSubmit={handleSubmit}>
        <label className={styles.label}>Current Password</label>
        <input
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          className={styles.input}
          required
        />

        <label className={styles.label}>New Password</label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          className={styles.input}
          required
        />

        <label className={styles.label}>Confirm New Password</label>
        <input
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          className={styles.input}
          required
        />

        <button type="submit" className={styles.btn} disabled={loading}>
          {loading ? 'Updating...' : 'Update Password'}
        </button>
      </form>

      {toast && (
        <Toast
          message={toast.message}
          type={toast.type === 'error' ? 'error' : 'success'}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  );
};

export default UpdatePasswordPage;
