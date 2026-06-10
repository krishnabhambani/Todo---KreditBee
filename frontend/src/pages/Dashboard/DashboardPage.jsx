import React, { useState, useEffect, useContext } from 'react';
import { useNavigate } from 'react-router-dom';
import API from '../../services/api';
import Toast from '../../components/Toast/Toast';
import { AuthContext } from '../../context/AuthContext';
import { 
  Search, 
  Trash2, 
  Edit, 
  Plus, 
  Calendar, 
  Clock, 
  ClipboardList,
  Users,
  User,
  AlertTriangle
} from 'lucide-react';
import styles from './DashboardPage.module.css';

const DashboardPage = () => {
  const { user } = useContext(AuthContext);
  const navigate = useNavigate();

  const [myGroups, setMyGroups] = useState([]);
  const [sharedGroups, setSharedGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);

  // Filters and search states
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all'); // all, active, completed, overdue, due-today, due-this-week
  const [sortBy, setSortBy] = useState('all'); // all, deadline, deadline-desc, updated, progress

  // Summary counts
  const [counts, setCounts] = useState({ overdue: 0, dueToday: 0, dueThisWeek: 0, upcoming: 0 });

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
    }, 300);
    return () => clearTimeout(timer);
  }, [search]);

  // Fetch groups
  useEffect(() => {
    fetchDashboardData();
    fetchSummaryCounts();
  }, [debouncedSearch, statusFilter, sortBy]);

  const fetchSummaryCounts = async () => {
    try {
      const myRes = await API.get('/groups');
      const sharedRes = await API.get('/shared-groups');
      const all = [...(myRes.data || []), ...(sharedRes.data || [])];
      
      setCounts({
        overdue: all.filter(g => g.health_status === 'OVERDUE').length,
        dueToday: all.filter(g => g.days_remaining === 0 && g.health_status !== 'COMPLETED').length,
        dueThisWeek: all.filter(g => g.days_remaining <= 7 && g.days_remaining >= 0 && g.health_status !== 'COMPLETED').length,
        upcoming: all.filter(g => g.days_remaining > 7 && g.health_status !== 'COMPLETED').length,
      });
    } catch (err) {
      console.error("Failed to fetch summary counts:", err);
    }
  };

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      let myParams = [];
      let sharedParams = [];

      if (debouncedSearch.trim()) {
        myParams.push(`search=${encodeURIComponent(debouncedSearch.trim())}`);
      }

      if (statusFilter !== 'all' && statusFilter !== 'upcoming') {
        myParams.push(`status=${statusFilter}`);
        sharedParams.push(`status=${statusFilter}`);
      } else if (statusFilter === 'upcoming') {
        myParams.push('status=active');
        sharedParams.push('status=active');
      }

      if (sortBy !== 'all' && sortBy !== 'progress') {
        myParams.push(`sort=${sortBy}`);
        sharedParams.push(`sort=${sortBy}`);
      }

      const myUrl = '/groups' + (myParams.length ? `?${myParams.join('&')}` : '');
      const sharedUrl = '/shared-groups' + (sharedParams.length ? `?${sharedParams.join('&')}` : '');

      const myRes = await API.get(myUrl);
      const sharedRes = await API.get(sharedUrl);

      let myData = myRes.data || [];
      let sharedData = sharedRes.data || [];

      // Filter in-memory for upcoming status (days_remaining > 7)
      if (statusFilter === 'upcoming') {
        myData = myData.filter(g => g.days_remaining > 7);
        sharedData = sharedData.filter(g => g.days_remaining > 7);
      }

      // Sort in-memory for progress
      if (sortBy === 'progress') {
        myData.sort((a, b) => b.progress - a.progress);
        sharedData.sort((a, b) => b.progress - a.progress);
      }

      setMyGroups(myData);
      setSharedGroups(sharedData);
    } catch (err) {
      setToast({ type: 'error', message: err.message });
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id, e) => {
    e.stopPropagation();
    if (!window.confirm('Are you sure you want to delete this group? All subtasks and sharing configurations will be permanently deleted.')) return;

    try {
      setMyGroups(myGroups.filter(g => g.id !== id));
      const response = await API.delete(`/groups/${id}`);
      if (response.success) {
        setToast({ type: 'success', message: 'Group deleted!' });
        fetchSummaryCounts();
      }
    } catch (err) {
      setToast({ type: 'error', message: err.message });
      fetchDashboardData(); // Revert
    }
  };

  const handleEdit = (id, e) => {
    e.stopPropagation();
    navigate(`/edit/${id}`);
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const getHealthBadgeStyle = (status) => {
    switch (status) {
      case 'COMPLETED':
        return styles.healthCompleted;
      case 'OVERDUE':
        return styles.healthOverdue;
      case 'AT_RISK':
        return styles.healthAtRisk;
      default:
        return styles.healthOnTrack;
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

  const filteredMyGroups = myGroups;
  const filteredSharedGroups = sharedGroups;

  return (
    <div className={styles.container}>
      {/* Upper header */}
      <div className={styles.dashboardHeader}>
        <div>
          <h1 className={styles.title}>Hello, {user?.name}!</h1>
          <p className={styles.subtitle}>Track, collaborate, and manage deadlines.</p>
        </div>
        <button className={styles.addBtn} onClick={() => navigate('/create')}>
          <Plus size={18} />
          <span>New Group</span>
        </button>
      </div>

      {/* Summary Widgets Row */}
      <div className={styles.summaryContainer}>
        <div 
          className={`${styles.summaryCard} ${statusFilter === 'overdue' ? styles.summaryCardActive : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'overdue' ? 'all' : 'overdue')}
        >
          <span className={styles.summaryValue} style={{ color: 'var(--color-danger)' }}>{counts.overdue}</span>
          <span className={styles.summaryLabel}>🔴 Overdue</span>
        </div>
        <div 
          className={`${styles.summaryCard} ${statusFilter === 'due-today' ? styles.summaryCardActive : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'due-today' ? 'all' : 'due-today')}
        >
          <span className={styles.summaryValue} style={{ color: 'var(--color-pending)' }}>{counts.dueToday}</span>
          <span className={styles.summaryLabel}>⏰ Due Today</span>
        </div>
        <div 
          className={`${styles.summaryCard} ${statusFilter === 'due-this-week' ? styles.summaryCardActive : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'due-this-week' ? 'all' : 'due-this-week')}
        >
          <span className={styles.summaryValue} style={{ color: 'var(--color-primary)' }}>{counts.dueThisWeek}</span>
          <span className={styles.summaryLabel}>📅 Due This Week</span>
        </div>
        <div 
          className={`${styles.summaryCard} ${statusFilter === 'upcoming' ? styles.summaryCardActive : ''}`}
          onClick={() => setStatusFilter(statusFilter === 'upcoming' ? 'all' : 'upcoming')}
        >
          <span className={styles.summaryValue} style={{ color: 'var(--color-success)' }}>{counts.upcoming}</span>
          <span className={styles.summaryLabel}>🔵 Upcoming</span>
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className={styles.filterSection}>
        <div className={styles.searchContainer}>
          <Search size={18} className={styles.searchIcon} />
          <input
            type="text"
            placeholder="Search groups..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flexWrap: 'wrap' }}>
          {/* Status filter selection */}
          <select 
            value={statusFilter} 
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{
              padding: '0.5rem 1rem',
              border: '1px solid var(--border-color)',
              borderRadius: '20px',
              fontSize: '0.875rem',
              fontWeight: '600',
              outline: 'none',
              backgroundColor: 'white',
              cursor: 'pointer',
              color: 'var(--text-primary)',
              boxShadow: 'var(--shadow-sm)'
            }}
          >
            <option value="all">All Statuses</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="overdue">Overdue Only</option>
            <option value="due-today">Due Today</option>
            <option value="due-this-week">Due This Week</option>
          </select>

          {/* Sorting selection */}
          <select 
            value={sortBy} 
            onChange={(e) => setSortBy(e.target.value)}
            style={{
              padding: '0.5rem 1rem',
              border: '1px solid var(--border-color)',
              borderRadius: '20px',
              fontSize: '0.875rem',
              fontWeight: '600',
              outline: 'none',
              backgroundColor: 'white',
              cursor: 'pointer',
              color: 'var(--text-primary)',
              boxShadow: 'var(--shadow-sm)'
            }}
          >
            <option value="all">Sort: Date Created</option>
            <option value="deadline">Sort: Nearest Deadline</option>
            <option value="deadline-desc">Sort: Furthest Deadline</option>
            <option value="updated">Sort: Recently Updated</option>
            <option value="progress">Sort: Progress %</option>
          </select>
        </div>
      </div>

      {/* Dashboard Lists */}
      {loading ? (
        <div className={styles.loadingContainer}>
          <div className="spinner"></div>
          <p>Loading dashboard data...</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          
          {/* Section 1: My Groups */}
          <div>
            <div className={styles.sectionHeader}>
              <h2 className={styles.sectionTitle}>My Groups ({filteredMyGroups.length})</h2>
            </div>
            
            {filteredMyGroups.length > 0 ? (
              <div className={styles.todoGrid}>
                {filteredMyGroups.map(group => {
                  const isOverdue = group.health_status === 'OVERDUE';
                  const daysLeft = group.days_remaining;
                  const overdueSubtasks = getOverdueSubtasksCount(group);
                  
                  return (
                    <div 
                      key={group.id} 
                      className={`${styles.todoCard} ${group.completed ? styles.completedCard : ''} ${styles.cardClickable}`}
                      onClick={() => navigate(`/groups/${group.id}`)}
                    >
                      <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '8px', marginBottom: '8px' }}>
                          <h3 className={styles.cardHeaderTitle} style={{ margin: 0 }}>
                            {group.title}
                          </h3>
                          <span className={`${styles.healthBadge} ${getHealthBadgeStyle(group.health_status)}`}>
                            {getHealthLabel(group.health_status)}
                          </span>
                        </div>
                        {group.description && (
                          <p className={styles.todoDesc}>
                            {group.description}
                          </p>
                        )}
                      </div>

                      {/* Group Badges Info */}
                      <div className={styles.badgesRow}>
                        <span className={styles.ownerBadge}>
                          <User size={12} />
                          <span>Owner: Me</span>
                        </span>
                        {group.member_count > 0 && (
                          <span className={styles.memberBadge}>
                            <Users size={12} />
                            <span>{group.member_count} {group.member_count === 1 ? 'Collaborator' : 'Collaborators'}</span>
                          </span>
                        )}
                      </div>

                      {/* Warnings banner container */}
                      {(daysLeft >= 0 && daysLeft <= 2 && group.health_status !== 'COMPLETED') || overdueSubtasks > 0 ? (
                        <div className={styles.warningsList}>
                          {daysLeft >= 0 && daysLeft <= 2 && group.health_status !== 'COMPLETED' && (
                            <div className={`${styles.warningTag} ${styles.warningAlert}`}>
                              <AlertTriangle size={12} />
                              <span>Deadline in {daysLeft} {daysLeft === 1 ? 'day' : 'days'}</span>
                            </div>
                          )}
                          {overdueSubtasks > 0 && (
                            <div className={`${styles.warningTag} ${styles.warningAlert}`}>
                              <AlertTriangle size={12} />
                              <span>{overdueSubtasks} {overdueSubtasks === 1 ? 'Subtask' : 'Subtasks'} Overdue</span>
                            </div>
                          )}
                        </div>
                      ) : null}

                      {/* Progress Tracker */}
                      <div className={styles.progressContainer}>
                        <div className={styles.progressText}>
                          <span>{group.completed_subtasks} / {group.total_subtasks} Completed</span>
                          <span>{Math.round(group.progress)}%</span>
                        </div>
                        <div className={styles.progressBarTrack}>
                          <div 
                            className={`${styles.progressBarFill} ${group.completed ? styles.progressBarFillCompleted : ''}`} 
                            style={{ width: `${group.progress}%` }}
                          />
                        </div>
                      </div>

                      {/* Card Footer */}
                      <div className={styles.cardFooter}>
                        <div className={styles.meta}>
                          <div className={`${styles.deadlineSection} ${isOverdue ? styles.overdue : ''}`}>
                            <Clock size={14} />
                            <span>
                              {group.due_date ? `Due: ${formatDate(group.due_date)}` : 'No deadline'}
                              {group.due_date && ` (${getDaysRemainingText(group)})`}
                            </span>
                          </div>
                        </div>

                        <div className={styles.actions}>
                          <button 
                            onClick={(e) => handleEdit(group.id, e)}
                            className={styles.editBtn}
                            title="Edit group"
                          >
                            <Edit size={16} />
                          </button>
                          <button 
                            onClick={(e) => handleDelete(group.id, e)}
                            className={styles.deleteBtn}
                            title="Delete group"
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className={styles.emptyState} style={{ padding: '30px' }}>
                <p>No owned groups matching current filters.</p>
              </div>
            )}
          </div>

          {/* Section 2: Shared With Me */}
          <div>
            <div className={styles.sectionHeader}>
              <h2 className={styles.sectionTitle}>Shared With Me ({filteredSharedGroups.length})</h2>
            </div>
            
            {filteredSharedGroups.length > 0 ? (
              <div className={styles.todoGrid}>
                {filteredSharedGroups.map(group => {
                  const isOverdue = group.health_status === 'OVERDUE';
                  const daysLeft = group.days_remaining;
                  const overdueSubtasks = getOverdueSubtasksCount(group);
                  const permissionLabel = group.user_permission === 'EDIT' ? 'Editor' : 'Viewer';
                  
                  return (
                    <div 
                      key={group.id} 
                      className={`${styles.todoCard} ${group.completed ? styles.completedCard : ''} ${styles.cardClickable}`}
                      onClick={() => navigate(`/groups/${group.id}`)}
                    >
                      <div>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '8px', marginBottom: '8px' }}>
                          <h3 className={styles.cardHeaderTitle} style={{ margin: 0 }}>
                            {group.title}
                          </h3>
                          <span className={`${styles.healthBadge} ${getHealthBadgeStyle(group.health_status)}`}>
                            {getHealthLabel(group.health_status)}
                          </span>
                        </div>
                        {group.description && (
                          <p className={styles.todoDesc}>
                            {group.description}
                          </p>
                        )}
                      </div>

                      {/* Group Badges Info */}
                      <div className={styles.badgesRow}>
                        <span className={styles.ownerBadge}>
                          <User size={12} />
                          <span>Owner: {group.owner?.name || 'Unknown'}</span>
                        </span>
                        <span className={styles.memberBadge} style={{ backgroundColor: 'var(--color-success-light)', color: 'var(--color-success)' }}>
                          <span>Access: {permissionLabel}</span>
                        </span>
                        {group.member_count > 0 && (
                          <span className={styles.memberBadge}>
                            <Users size={12} />
                            <span>{group.member_count} {group.member_count === 1 ? 'Collaborator' : 'Collaborators'}</span>
                          </span>
                        )}
                      </div>

                      {/* Warnings banner container */}
                      {(daysLeft >= 0 && daysLeft <= 2 && group.health_status !== 'COMPLETED') || overdueSubtasks > 0 ? (
                        <div className={styles.warningsList}>
                          {daysLeft >= 0 && daysLeft <= 2 && group.health_status !== 'COMPLETED' && (
                            <div className={`${styles.warningTag} ${styles.warningAlert}`}>
                              <AlertTriangle size={12} />
                              <span>Deadline in {daysLeft} {daysLeft === 1 ? 'day' : 'days'}</span>
                            </div>
                          )}
                          {overdueSubtasks > 0 && (
                            <div className={`${styles.warningTag} ${styles.warningAlert}`}>
                              <AlertTriangle size={12} />
                              <span>{overdueSubtasks} {overdueSubtasks === 1 ? 'Subtask' : 'Subtasks'} Overdue</span>
                            </div>
                          )}
                        </div>
                      ) : null}

                      {/* Progress Tracker */}
                      <div className={styles.progressContainer}>
                        <div className={styles.progressText}>
                          <span>{group.completed_subtasks} / {group.total_subtasks} Completed</span>
                          <span>{Math.round(group.progress)}%</span>
                        </div>
                        <div className={styles.progressBarTrack}>
                          <div 
                            className={`${styles.progressBarFill} ${group.completed ? styles.progressBarFillCompleted : ''}`} 
                            style={{ width: `${group.progress}%` }}
                          />
                        </div>
                      </div>

                      {/* Card Footer */}
                      <div className={styles.cardFooter}>
                        <div className={styles.meta}>
                          <div className={`${styles.deadlineSection} ${isOverdue ? styles.overdue : ''}`}>
                            <Clock size={14} />
                            <span>
                              {group.due_date ? `Due: ${formatDate(group.due_date)}` : 'No deadline'}
                              {group.due_date && ` (${getDaysRemainingText(group)})`}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className={styles.emptyState} style={{ padding: '30px' }}>
                <p>No shared groups matching current filters.</p>
              </div>
            )}
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

export default DashboardPage;
