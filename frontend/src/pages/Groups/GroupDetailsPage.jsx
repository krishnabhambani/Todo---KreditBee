import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import API from '../../services/api';
import Toast from '../../components/Toast/Toast';
import { 
  ArrowLeft, 
  Plus, 
  Trash2, 
  Edit, 
  Check, 
  Calendar, 
  Clock, 
  Save, 
  X,
  AlertCircle,
  AlertTriangle,
  Users,
  User
} from 'lucide-react';
import styles from './GroupDetailsPage.module.css';

const GroupDetailsPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [group, setGroup] = useState(null);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);
  const [filter, setFilter] = useState('All'); // All, Active, Completed

  // Subtask creation state
  const [isAddingSubtask, setIsAddingSubtask] = useState(false);
  const [newSubtask, setNewSubtask] = useState({ title: '', description: '', due_date: '' });

  // Subtask editing state
  const [editingSubtaskId, setEditingSubtaskId] = useState(null);
  const [editSubtaskData, setEditSubtaskData] = useState({ title: '', description: '', due_date: '' });

  // Sharing states
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);
  const [collaborators, setCollaborators] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [sharingEmail, setSharingEmail] = useState('');
  const [sharePermission, setSharePermission] = useState('VIEW'); // VIEW or EDIT

  useEffect(() => {
    fetchGroupDetails();
  }, [id]);

  const fetchGroupDetails = async () => {
    setLoading(true);
    try {
      const response = await API.get(`/groups/${id}`);
      if (response.success && response.data) {
        setGroup(response.data);
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
      setTimeout(() => navigate('/'), 1500);
    } finally {
      setLoading(false);
    }
  };

  const fetchCollaborators = async () => {
    try {
      const response = await API.get(`/groups/${id}/members`);
      if (response.success && response.data) {
        setCollaborators(response.data || []);
      }
    } catch (err) {
      console.error("Failed to load collaborators:", err);
    }
  };

  const handleOpenShareModal = () => {
    setIsShareModalOpen(true);
    fetchCollaborators();
  };

  const handleCloseShareModal = () => {
    setIsShareModalOpen(false);
    setSearchQuery('');
    setSearchResults([]);
    setSharingEmail('');
    setSharePermission('VIEW');
  };

  const handleSearchUsers = async (query) => {
    setSearchQuery(query);
    setSharingEmail(query); // Sync typed text as fallback
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    try {
      const response = await API.get(`/users?search=${encodeURIComponent(query)}`);
      if (response.success && response.data) {
        setSearchResults(response.data || []);
      }
    } catch (err) {
      console.error("User search failed:", err);
    }
  };

  const handleSelectUser = (user) => {
    setSharingEmail(user.email);
    setSearchQuery(user.name + " (" + user.email + ")");
    setSearchResults([]);
  };

  const handleShareGroupSubmit = async (e) => {
    e.preventDefault();
    if (!sharingEmail.trim()) {
      setToast({ type: 'error', message: 'Email address is required' });
      return;
    }
    try {
      const response = await API.post(`/groups/${id}/share`, {
        email: sharingEmail.trim(),
        permission: sharePermission
      });
      if (response.success) {
        setToast({ type: 'success', message: 'Group shared successfully!' });
        setSharingEmail('');
        setSearchQuery('');
        fetchCollaborators();
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
    }
  };

  const handleRevokeShare = async (targetUserId) => {
    if (!window.confirm("Are you sure you want to revoke access for this user?")) return;
    try {
      const response = await API.delete(`/groups/${id}/share/${targetUserId}`);
      if (response.success) {
        setToast({ type: 'success', message: 'Access revoked successfully' });
        fetchCollaborators();
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
    }
  };

  const formatDateForInput = (dateStr) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '';
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  // Recalculates total, completed and progress percentages locally for instant reactivity
  const updateLocalGroupState = (updatedSubtasks) => {
    const total = updatedSubtasks.length;
    const completed = updatedSubtasks.filter(s => s.completed).length;
    const progress = total === 0 ? 0 : (completed / total) * 100;
    
    setGroup(prev => {
      const updatedGroup = {
        ...prev,
        subtasks: updatedSubtasks,
        total_subtasks: total,
        completed_subtasks: completed,
        progress: progress,
        completed: total > 0 && completed === total
      };

      // Recalculate health status locally
      const now = new Date();
      if (updatedGroup.completed) {
        updatedGroup.health_status = 'COMPLETED';
      } else if (updatedGroup.due_date && new Date(updatedGroup.due_date) < now) {
        updatedGroup.health_status = 'OVERDUE';
      } else if (updatedGroup.due_date) {
        const diff = new Date(updatedGroup.due_date) - now;
        const days = diff / (1000 * 60 * 60 * 24);
        if (days <= 3 && updatedGroup.progress < 75) {
          updatedGroup.health_status = 'AT_RISK';
        } else if (days <= 7 && updatedGroup.progress < 40) {
          updatedGroup.health_status = 'AT_RISK';
        } else {
          updatedGroup.health_status = 'ON_TRACK';
        }
      } else {
        updatedGroup.health_status = 'ON_TRACK';
      }

      return updatedGroup;
    });
  };

  // Subtask Handlers
  const handleToggleComplete = async (subtaskId) => {
    try {
      // Optimistic update
      const updatedSubtasks = group.subtasks.map(s => 
        s.id === subtaskId ? { ...s, completed: !s.completed } : s
      );
      updateLocalGroupState(updatedSubtasks);

      const response = await API.patch(`/tasks/${subtaskId}/complete`);
      if (response.success) {
        setToast({ type: 'success', message: 'Subtask status updated' });
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
      fetchGroupDetails(); // Revert on failure
    }
  };

  const handleDeleteSubtask = async (subtaskId) => {
    if (!window.confirm('Delete this subtask?')) return;

    try {
      // Optimistic update
      const updatedSubtasks = group.subtasks.filter(s => s.id !== subtaskId);
      updateLocalGroupState(updatedSubtasks);

      const response = await API.delete(`/tasks/${subtaskId}`);
      if (response.success) {
        setToast({ type: 'success', message: 'Subtask deleted' });
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
      fetchGroupDetails(); // Revert on failure
    }
  };

  const handleAddSubtask = async (e) => {
    e.preventDefault();
    if (!newSubtask.title.trim()) {
      setToast({ type: 'error', message: 'Subtask title is required' });
      return;
    }

    try {
      const payload = {
        title: newSubtask.title.trim(),
        description: newSubtask.description.trim(),
        due_date: newSubtask.due_date ? new Date(newSubtask.due_date).toISOString() : null
      };

      const response = await API.post(`/groups/${id}/tasks`, payload);
      if (response.success && response.data) {
        const updatedSubtasks = [...(group.subtasks || []), response.data];
        updateLocalGroupState(updatedSubtasks);
        
        // Reset state
        setNewSubtask({ title: '', description: '', due_date: '' });
        setIsAddingSubtask(false);
        setToast({ type: 'success', message: 'Subtask added!' });
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
    }
  };

  const handleStartEdit = (subtask) => {
    setEditingSubtaskId(subtask.id);
    setEditSubtaskData({
      title: subtask.title,
      description: subtask.description || '',
      due_date: formatDateForInput(subtask.due_date)
    });
  };

  const handleUpdateSubtask = async (e, subtaskId) => {
    e.preventDefault();
    if (!editSubtaskData.title.trim()) {
      setToast({ type: 'error', message: 'Title is required' });
      return;
    }

    try {
      const payload = {
        title: editSubtaskData.title.trim(),
        description: editSubtaskData.description.trim(),
        due_date: editSubtaskData.due_date ? new Date(editSubtaskData.due_date).toISOString() : null
      };

      const response = await API.put(`/tasks/${subtaskId}`, payload);
      if (response.success && response.data) {
        const updatedSubtasks = group.subtasks.map(s => 
          s.id === subtaskId ? response.data : s
        );
        updateLocalGroupState(updatedSubtasks);
        setEditingSubtaskId(null);
        setToast({ type: 'success', message: 'Subtask updated!' });
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
    }
  };

  // Helper to determine if subtask is overdue
  const isOverdue = (subtask) => {
    if (!subtask.due_date || subtask.completed) return false;
    return new Date(subtask.due_date) < new Date();
  };

  // Filtering list
  const getFilteredSubtasks = () => {
    const list = group.subtasks || [];
    if (filter === 'Completed') {
      return list.filter(s => s.completed);
    } else if (filter === 'Active') {
      return list.filter(s => !s.completed);
    }
    return list;
  };

  const getHealthBadgeStyle = (status) => {
    switch (status) {
      case 'COMPLETED': return styles.healthCompleted;
      case 'OVERDUE': return styles.healthOverdue;
      case 'AT_RISK': return styles.healthAtRisk;
      default: return styles.healthOnTrack;
    }
  };

  const getHealthLabel = (status) => {
    switch (status) {
      case 'COMPLETED': return '🔵 Completed';
      case 'OVERDUE': return '🔴 Overdue';
      case 'AT_RISK': return '🟡 At Risk';
      default: return '🟢 On Track';
    }
  };

  const getDaysRemainingText = (group) => {
    if (group.health_status === 'COMPLETED') return 'Completed';
    if (group.days_remaining === 9999) return 'No deadline';
    if (group.days_remaining < 0) return `Overdue by ${Math.abs(group.days_remaining)} days`;
    if (group.days_remaining === 0) return 'Due Today';
    if (group.days_remaining === 1) return '1 day remaining';
    return `${group.days_remaining} days remaining`;
  };

  const getOverdueSubtasksCount = (group) => {
    if (!group.subtasks) return 0;
    const now = new Date();
    return group.subtasks.filter(s => !s.completed && s.due_date && new Date(s.due_date) < now).length;
  };

  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <div className="spinner"></div>
        <p>Loading group details...</p>
      </div>
    );
  }

  if (!group) {
    return (
      <div className={styles.emptyState}>
        <h3>Group not found</h3>
        <button onClick={() => navigate('/')} className={styles.backBtn}>
          <ArrowLeft size={16} />
          <span>Return Dashboard</span>
        </button>
      </div>
    );
  }

  const groupDeadlineOverdue = group.due_date && new Date(group.due_date) < new Date() && !group.completed;
  const filteredSubtasks = getFilteredSubtasks();

  const userPermission = group.user_permission || 'VIEW';
  const canToggle = userPermission === 'OWNER' || userPermission === 'EDIT';
  const canDeleteSubtask = userPermission === 'OWNER';
  const canManageSharing = userPermission === 'OWNER';

  const daysLeft = group.days_remaining;
  const overdueSubtasks = getOverdueSubtasksCount(group);

  return (
    <div className={styles.container}>
      {/* Header action buttons row */}
      <div className={styles.headerActionsRow}>
        <button onClick={() => navigate('/')} className={styles.backBtn}>
          <ArrowLeft size={18} />
          <span>Back to Dashboard</span>
        </button>
        {canManageSharing && (
          <button onClick={handleOpenShareModal} className={styles.shareBtn}>
            <Users size={16} />
            <span>Share Group</span>
          </button>
        )}
      </div>

      {/* Group Info Header Card */}
      <div className={styles.headerCard}>
        <div className={styles.groupInfo}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '8px', flexWrap: 'wrap' }}>
            <h1 className={styles.groupTitle} style={{ margin: 0 }}>{group.title}</h1>
            <span className={`${styles.healthBadge} ${getHealthBadgeStyle(group.health_status)}`}>
              {getHealthLabel(group.health_status)}
            </span>
          </div>
          {group.description && <p className={styles.groupDesc}>{group.description}</p>}
        </div>

        {/* Warnings Banner in details */}
        {(daysLeft >= 0 && daysLeft <= 2 && group.health_status !== 'COMPLETED') || overdueSubtasks > 0 ? (
          <div className={styles.warningsList} style={{ borderTop: '1px solid var(--border-color)', paddingTop: '12px' }}>
            {daysLeft >= 0 && daysLeft <= 2 && group.health_status !== 'COMPLETED' && (
              <div className={`${styles.warningTag} ${styles.warningAlert}`}>
                <AlertTriangle size={14} />
                <span>⚠ Project Deadline Approaching: {daysLeft} {daysLeft === 1 ? 'day' : 'days'} remaining!</span>
              </div>
            )}
            {overdueSubtasks > 0 && (
              <div className={`${styles.warningTag} ${styles.warningAlert}`}>
                <AlertTriangle size={14} />
                <span>⚠ Checklist Alert: {overdueSubtasks} overdue subtask {overdueSubtasks === 1 ? 'item' : 'items'} detected!</span>
              </div>
            )}
          </div>
        ) : null}

        <div className={styles.metaInfo}>
          {group.due_date ? (
            <div className={`${styles.metaItem} ${groupDeadlineOverdue ? styles.overdue : ''}`}>
              <Clock size={16} />
              <span>Deadline: {formatDate(group.due_date)} {groupDeadlineOverdue && '(Overdue)'} ({getDaysRemainingText(group)})</span>
            </div>
          ) : (
            <div className={styles.metaItem}>
              <Calendar size={16} />
              <span>No group deadline set</span>
            </div>
          )}
          
          <div className={styles.metaItem}>
            <span>Subtasks: {group.completed_subtasks} / {group.total_subtasks}</span>
          </div>

          <div className={styles.metaItem}>
            <span className={styles.permissionBadge} style={{
              backgroundColor: userPermission === 'OWNER' ? 'var(--color-primary-light)' : userPermission === 'EDIT' ? 'var(--color-success-light)' : '#f1f5f9',
              color: userPermission === 'OWNER' ? 'var(--color-primary)' : userPermission === 'EDIT' ? 'var(--color-success)' : 'var(--text-secondary)',
              fontSize: '12px'
            }}>
              Role: {userPermission}
            </span>
          </div>
        </div>

        {/* Overall Group Progress Bar */}
        <div className={styles.progressContainer}>
          <div className={styles.progressText}>
            <span>Overall Progress</span>
            <span>{Math.round(group.progress)}%</span>
          </div>
          <div className={styles.progressBarTrack}>
            <div 
              className={`${styles.progressBarFill} ${group.completed ? styles.progressBarFillCompleted : ''}`} 
              style={{ width: `${group.progress}%` }}
            />
          </div>
        </div>
      </div>

      {/* Subtasks Section */}
      <div className={styles.subtasksSection}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Subtasks Checklist</h2>
          
          <div className={styles.filters}>
            {['All', 'Active', 'Completed'].map(f => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`${styles.filterBtn} ${filter === f ? styles.activeFilter : ''}`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>

        {/* Subtask Listing */}
        <div className={styles.subtaskList}>
          {filteredSubtasks.length > 0 ? (
            filteredSubtasks.map(subtask => {
              const subtaskOverdue = isOverdue(subtask);
              const isEditing = editingSubtaskId === subtask.id;

              if (isEditing) {
                return (
                  <div key={subtask.id} className={styles.formCard}>
                    <form onSubmit={(e) => handleUpdateSubtask(e, subtask.id)} className={styles.inlineForm}>
                      <input
                        type="text"
                        placeholder="Subtask title"
                        value={editSubtaskData.title}
                        onChange={(e) => setEditSubtaskData({ ...editSubtaskData, title: e.target.value })}
                        required
                        autoFocus
                      />
                      <textarea
                        placeholder="Subtask description (optional)"
                        value={editSubtaskData.description}
                        onChange={(e) => setEditSubtaskData({ ...editSubtaskData, description: e.target.value })}
                        rows={2}
                      />
                      <div className={styles.formRow}>
                        <div>
                          <label>Due Date</label>
                          <input
                            type="date"
                            value={editSubtaskData.due_date}
                            onChange={(e) => setEditSubtaskData({ ...editSubtaskData, due_date: e.target.value })}
                          />
                        </div>
                      </div>
                      <div className={styles.formActions}>
                        <button 
                          type="button" 
                          onClick={() => setEditingSubtaskId(null)}
                          className={styles.cancelBtn}
                        >
                          Cancel
                        </button>
                        <button type="submit" className={styles.submitBtn}>
                          <Save size={16} />
                          <span>Save</span>
                        </button>
                      </div>
                    </form>
                  </div>
                );
              }

              return (
                <div 
                  key={subtask.id} 
                  className={`${styles.subtaskItem} ${subtask.completed ? styles.subtaskItemCompleted : ''} ${subtaskOverdue ? styles.subtaskItemOverdue : ''}`}
                >
                  <div className={styles.subtaskHeader}>
                    <div className={styles.checkboxContainer}>
                      <button 
                        onClick={() => canToggle && handleToggleComplete(subtask.id)}
                        disabled={!canToggle}
                        style={{ cursor: canToggle ? 'pointer' : 'not-allowed' }}
                        className={`${styles.checkbox} ${subtask.completed ? styles.checked : ''}`}
                      >
                        {subtask.completed && <Check size={14} />}
                      </button>
                      <span className={`${styles.subtaskTitle} ${subtask.completed ? styles.lineThrough : ''}`}>
                        {subtask.title}
                      </span>
                    </div>

                    <div className={styles.actions}>
                      {canToggle && (
                        <button 
                          onClick={() => handleStartEdit(subtask)}
                          className={styles.editBtn}
                          title="Edit subtask"
                        >
                          <Edit size={15} />
                        </button>
                      )}
                      {canDeleteSubtask && (
                        <button 
                          onClick={() => handleDeleteSubtask(subtask.id)}
                          className={styles.deleteBtn}
                          title="Delete subtask"
                        >
                          <Trash2 size={15} />
                        </button>
                      )}
                    </div>
                  </div>

                  {subtask.description && (
                    <p className={`${styles.subtaskDesc} ${subtask.completed ? styles.lineThrough : ''}`}>
                      {subtask.description}
                    </p>
                  )}

                  <div className={styles.subtaskFooter}>
                    {subtask.due_date ? (
                      <div className={`${styles.subtaskDueDate} ${subtaskOverdue ? styles.overdue : ''}`}>
                        {subtaskOverdue ? <AlertCircle size={13} /> : <Calendar size={13} />}
                        <span>Due: {formatDate(subtask.due_date)} {subtaskOverdue && '(Overdue)'}</span>
                      </div>
                    ) : (
                      <span>No deadline</span>
                    )}
                  </div>
                </div>
              );
            })
          ) : (
            <div className={styles.emptyState}>
              <p>No subtasks found in this filter.</p>
            </div>
          )}

          {/* Add Subtask Form / Trigger */}
          {canToggle && (
            isAddingSubtask ? (
              <div className={styles.formCard}>
                <form onSubmit={handleAddSubtask} className={styles.inlineForm}>
                  <h3 style={{ fontSize: '15px', fontWeight: '700', color: 'var(--text-primary)', marginBottom: '4px' }}>
                    Add New Subtask
                  </h3>
                  <input
                    type="text"
                    placeholder="Task title (e.g. Design landing page)"
                    value={newSubtask.title}
                    onChange={(e) => setNewSubtask({ ...newSubtask, title: e.target.value })}
                    required
                    autoFocus
                  />
                  <textarea
                    placeholder="Description (optional)"
                    value={newSubtask.description}
                    onChange={(e) => setNewSubtask({ ...newSubtask, description: e.target.value })}
                    rows={2}
                  />
                  <div className={styles.formRow}>
                    <div>
                      <label>Due Date</label>
                      <input
                        type="date"
                        value={newSubtask.due_date}
                        onChange={(e) => setNewSubtask({ ...newSubtask, due_date: e.target.value })}
                      />
                    </div>
                  </div>
                  <div className={styles.formActions}>
                    <button 
                      type="button" 
                      onClick={() => setIsAddingSubtask(false)}
                      className={styles.cancelBtn}
                    >
                      Cancel
                    </button>
                    <button type="submit" className={styles.submitBtn}>
                      <Plus size={16} />
                      <span>Add Task</span>
                    </button>
                  </div>
                </form>
              </div>
            ) : (
              <button 
                onClick={() => setIsAddingSubtask(true)}
                className={styles.addTriggerBtn}
              >
                <Plus size={18} />
                <span>Add New Subtask</span>
              </button>
            )
          )}
        </div>
      </div>

      {/* Share Modal */}
      {isShareModalOpen && (
        <div className={styles.modalOverlay} onClick={handleCloseShareModal}>
          <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2>Share "{group.title}"</h2>
              <button onClick={handleCloseShareModal} className={styles.closeBtn}>
                <X size={20} />
              </button>
            </div>

            <div className={styles.modalBody}>
              {/* User share form */}
              <form onSubmit={handleShareGroupSubmit} className={styles.shareForm}>
                <div className={styles.shareInputsRow}>
                  <div className={styles.searchWrapper}>
                    <label>Collaborator Email / Name</label>
                    <input
                      type="text"
                      placeholder="Type email or name to search..."
                      value={searchQuery}
                      onChange={(e) => handleSearchUsers(e.target.value)}
                      required
                    />
                    
                    {searchResults.length > 0 && (
                      <div className={styles.searchResultsList}>
                        {searchResults.map(u => (
                          <div 
                            key={u.id} 
                            className={styles.searchResultItem}
                            onClick={() => handleSelectUser(u)}
                          >
                            <span className={styles.resultName}>{u.name}</span>
                            <span className={styles.resultEmail}>{u.email}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  <div>
                    <label>Permission</label>
                    <select 
                      value={sharePermission} 
                      onChange={(e) => setSharePermission(e.target.value)}
                      className={styles.selectInput}
                    >
                      <option value="VIEW">Viewer</option>
                      <option value="EDIT">Editor</option>
                    </select>
                  </div>
                </div>

                <button 
                  type="submit" 
                  className={styles.submitBtn} 
                  style={{ width: '100%', padding: '12px' }}
                >
                  <Plus size={16} />
                  <span>Add Collaborator</span>
                </button>
              </form>

              {/* Collaborators list */}
              <h3 style={{ fontSize: '15px', fontWeight: '700', color: 'var(--text-primary)', marginTop: '24px', marginBottom: '8px' }}>
                Current Collaborators ({collaborators.length})
              </h3>
              
              <div className={styles.collaboratorsList}>
                {collaborators.length > 0 ? (
                  collaborators.map(member => (
                    <div key={member.id} className={styles.collaboratorItem}>
                      <div className={styles.collabInfo}>
                        <span className={styles.collabName}>{member.shared_with?.name}</span>
                        <span className={styles.collabEmail}>{member.shared_with?.email}</span>
                      </div>

                      <div className={styles.collabActions}>
                        <span className={`${styles.permissionBadge} ${member.permission === 'EDIT' ? styles.badgeEdit : styles.badgeView}`}>
                          {member.permission === 'EDIT' ? 'Editor' : 'Viewer'}
                        </span>
                        
                        {canManageSharing && (
                          <button 
                            onClick={() => handleRevokeShare(member.shared_with_user_id)}
                            className={styles.deleteBtn}
                            style={{ padding: '4px' }}
                            title="Revoke access"
                          >
                            <Trash2 size={15} />
                          </button>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <p style={{ fontSize: '13px', color: 'var(--text-secondary)', textAlign: 'center', padding: '12px' }}>
                    This group isn't shared with anyone yet.
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

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

export default GroupDetailsPage;
