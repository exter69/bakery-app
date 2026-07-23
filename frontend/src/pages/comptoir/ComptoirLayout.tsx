import { Outlet } from 'react-router-dom';
import { ComptoirNav } from './ComptoirNav';
import './ComptoirLayout.css';

export default function ComptoirLayout() {
  return (
    <div className="comptoir-layout">
      <ComptoirNav />
      <main className="comptoir-main">
        <Outlet />
      </main>
    </div>
  );
}
