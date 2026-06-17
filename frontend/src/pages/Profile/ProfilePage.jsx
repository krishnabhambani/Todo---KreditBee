import React, { useContext } from 'react';
import { AuthContext } from '../../context/AuthContext';
import styles from './ProfilePage.module.css';
import { User, Mail, Calendar } from 'lucide-react';

const ProfilePage = () => {
  const { user } = useContext(AuthContext);

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>User Profile</h1>
      
      <div className={styles.card}>
        <div className={styles.avatarLarge}>
          {user?.name?.charAt(0).toUpperCase()}
        </div>
        
        <div className={styles.profileDetails}>
          <div className={styles.infoRow}>
            <User className={styles.icon} size={20} />
            <div className={styles.infoContent}>
              <span className={styles.label}>Full Name</span>
              <span className={styles.value}>{user?.name || 'N/A'}</span>
            </div>
          </div>

          <div className={styles.infoRow}>
            <Mail className={styles.icon} size={20} />
            <div className={styles.infoContent}>
              <span className={styles.label}>Email Address</span>
              <span className={styles.value}>{user?.email || 'N/A'}</span>
            </div>
          </div>

          {/* <div className={styles.infoRow}>
            <Calendar className={styles.icon} size={20} />
            <div className={styles.infoContent}>
              <span className={styles.label}>Account ID</span>
              <span className={styles.value}>User #{user?.id || 'N/A'}</span>
            </div>
          </div> */}
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
