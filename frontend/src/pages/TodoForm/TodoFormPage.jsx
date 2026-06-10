import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import API from '../../services/api';
import Toast from '../../components/Toast/Toast';
import { Save, ArrowLeft } from 'lucide-react';
import styles from './TodoFormPage.module.css';

const TodoFormPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = !!id;

  const [formData, setFormData] = useState({ title: '', description: '', due_date: '' });
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [toast, setToast] = useState(null);

  useEffect(() => {
    if (isEdit) {
      fetchGroup();
    }
  }, [id]);

  const formatDateForInput = (dateStr) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  const fetchGroup = async () => {
    setFetching(true);
    try {
      const response = await API.get(`/groups/${id}`);
      if (response.success && response.data) {
        setFormData({
          title: response.data.title,
          description: response.data.description || '',
          due_date: formatDateForInput(response.data.due_date),
        });
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
      setTimeout(() => navigate('/'), 1500);
    } finally {
      setFetching(false);
    }
  };

  const validate = () => {
    const tempErrors = {};
    if (!formData.title.trim()) {
      tempErrors.title = 'Title is required';
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
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;

    setLoading(true);
    try {
      const payload = {
        title: formData.title.trim(),
        description: formData.description.trim(),
        due_date: formData.due_date ? new Date(formData.due_date).toISOString() : null,
      };

      let response;
      if (isEdit) {
        response = await API.put(`/groups/${id}`, payload);
      } else {
        response = await API.post('/groups', payload);
      }

      if (response.success) {
        setToast({ type: 'success', message: isEdit ? 'Group updated!' : 'Group created!' });
        setTimeout(() => navigate('/'), 1000);
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
    } finally {
      setLoading(false);
    }
  };

  if (fetching) {
    return (
      <div className={styles.loadingContainer}>
        <div className="spinner"></div>
        <p>Loading details...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <button onClick={() => navigate(-1)} className={styles.backBtn}>
        <ArrowLeft size={18} />
        <span>Back</span>
      </button>

      <h1 className={styles.title}>{isEdit ? 'Edit Group' : 'Create New Group'}</h1>

      <div className={styles.card}>
        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.inputGroup}>
            <label>Group Title <span className={styles.required}>*</span></label>
            <input
              type="text"
              name="title"
              placeholder="What is the group/project title?"
              value={formData.title}
              onChange={handleChange}
              className={errors.title ? styles.inputError : ''}
              disabled={loading}
              autoFocus
            />
            {errors.title && <span className={styles.errorText}>{errors.title}</span>}
          </div>

          <div className={styles.inputGroup}>
            <label>Description</label>
            <textarea
              name="description"
              placeholder="Add details about this group/project..."
              value={formData.description}
              onChange={handleChange}
              rows={5}
              disabled={loading}
            />
          </div>

          <div className={styles.inputGroup}>
            <label>Due Date</label>
            <input
              type="date"
              name="due_date"
              value={formData.due_date}
              onChange={handleChange}
              disabled={loading}
            />
          </div>

          <button type="submit" className={styles.submitBtn} disabled={loading}>
            {loading ? (
              <span className="spinner" style={{ width: '20px', height: '20px' }}></span>
            ) : (
              <>
                <Save size={18} />
                <span>{isEdit ? 'Save Changes' : 'Create Group'}</span>
              </>
            )}
          </button>
        </form>
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

export default TodoFormPage;
