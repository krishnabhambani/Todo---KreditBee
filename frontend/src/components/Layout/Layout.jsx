import React, { useContext, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import { 
  CheckSquare, 
  LayoutDashboard, 
  PlusCircle, 
  User, 
  LogOut, 
  Menu, 
  X,
  FolderOpen
} from 'lucide-react';
import styles from './Layout.module.css';

const Layout = ({ children }) => {
  const { user, logout } = useContext(AuthContext);
  const location = useLocation();
  const navigate = useNavigate();
  const [mobileOpen, setMobileOpen] = useState(false);

  const menuItems = [
    { name: 'Dashboard', path: '/', icon: <LayoutDashboard size={20} /> },
    { name: 'Create Group', path: '/create', icon: <PlusCircle size={20} /> },
    { name: 'Profile', path: '/profile', icon: <User size={20} /> },
  ];

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const toggleMobile = () => setMobileOpen(!mobileOpen);

  return (
    <div className={styles.container}>
      {/* Sidebar - Desktop */}
      <aside className={`${styles.sidebar} ${mobileOpen ? styles.sidebarOpen : ''}`}>
        <div className={styles.sidebarHeader}>
          <div className={styles.logo}>
            <CheckSquare size={24} className={styles.logoIcon} />
            <span>TodoApp</span>
          </div>
          <button className={styles.closeMobileBtn} onClick={toggleMobile}>
            <X size={20} />
          </button>
        </div>

        <nav className={styles.nav}>
          {menuItems.map((item) => (
            <Link
              key={item.name}
              to={item.path}
              className={`${styles.navLink} ${location.pathname === item.path ? styles.active : ''}`}
              onClick={() => setMobileOpen(false)}
            >
              {item.icon}
              <span>{item.name}</span>
            </Link>
          ))}
        </nav>

        <div className={styles.sidebarFooter}>
          <button className={styles.logoutBtn} onClick={handleLogout}>
            <LogOut size={20} />
            <span>Logout</span>
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className={styles.mainContent}>
        {/* Navbar */}
        <header className={styles.navbar}>
          <button className={styles.menuBtn} onClick={toggleMobile}>
            <Menu size={24} />
          </button>
          
          <div className={styles.navbarRight}>
            <div className={styles.userInfo}>
              <span className={styles.userName}>{user?.name}</span>
              <div className={styles.avatar}>
                {user?.name?.charAt(0).toUpperCase()}
              </div>
            </div>
          </div>
        </header>

        {/* Dynamic Page Content */}
        <main className={styles.contentBody}>
          {children}
        </main>
      </div>

      {/* Mobile Backdrop overlay */}
      {mobileOpen && <div className={styles.backdrop} onClick={toggleMobile} />}
    </div>
  );
};

export default Layout;
