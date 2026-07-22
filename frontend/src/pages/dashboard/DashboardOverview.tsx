import { useState } from 'react';
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Legend,
} from 'recharts';
import './DashboardOverview.css';

// --- Fake data ---

const monthlyData = [
  { month: 'Jan', orders: 42, deliveries: 38 },
  { month: 'Feb', orders: 55, deliveries: 48 },
  { month: 'Mar', orders: 63, deliveries: 57 },
  { month: 'Apr', orders: 78, deliveries: 70 },
  { month: 'May', orders: 85, deliveries: 76 },
  { month: 'Jun', orders: 92, deliveries: 84 },
  { month: 'Jul', orders: 75, deliveries: 68 },
  { month: 'Aug', orders: 60, deliveries: 52 },
  { month: 'Sep', orders: 88, deliveries: 80 },
  { month: 'Oct', orders: 95, deliveries: 87 },
  { month: 'Nov', orders: 110, deliveries: 98 },
  { month: 'Dec', orders: 120, deliveries: 105 },
];

const productSales = [
  { name: 'Croissant', sold: 340 },
  { name: 'Baguette Tradition', sold: 280 },
  { name: 'Pain au Chocolat', sold: 245 },
  { name: 'Tarte aux Pommes', sold: 120 },
  { name: 'Éclair au Café', sold: 95 },
  { name: 'Pain de Campagne', sold: 88 },
];

const peakHours = [
  { hour: '6h', orders: 8 },
  { hour: '7h', orders: 35 },
  { hour: '8h', orders: 52 },
  { hour: '9h', orders: 28 },
  { hour: '10h', orders: 15 },
  { hour: '11h', orders: 12 },
  { hour: '12h', orders: 18 },
  { hour: '13h', orders: 10 },
  { hour: '14h', orders: 8 },
  { hour: '15h', orders: 6 },
  { hour: '16h', orders: 12 },
  { hour: '17h', orders: 20 },
  { hour: '18h', orders: 15 },
];

const fulfillment = [
  { name: 'Delivered', value: 680, color: '#22c55e' },
  { name: 'Cancelled', value: 45, color: '#ef4444' },
  { name: 'Pending', value: 32, color: '#f59e0b' },
];

const topCustomers = [
  { name: 'Alice M.', orders: 28 },
  { name: 'Pierre D.', orders: 22 },
  { name: 'Sophie L.', orders: 19 },
  { name: 'Marc B.', orders: 15 },
  { name: 'Julie R.', orders: 12 },
];

type Tab = 'orders' | 'products' | 'customers';

export default function DashboardOverview() {
  const [tab, setTab] = useState<Tab>('orders');

  return (
    <div className="overview">
      <h1 className="overview__title">Dashboard Overview</h1>

      {/* Stat cards */}
      <div className="overview__stats">
        <div className="stat-card">
          <span className="stat-card__label">Total Revenue</span>
          <span className="stat-card__value">€12,450</span>
          <span className="stat-card__delta stat-card__delta--positive">+12% vs last month</span>
        </div>
        <div className="stat-card">
          <span className="stat-card__label">Total Orders</span>
          <span className="stat-card__value">757</span>
          <span className="stat-card__delta">All time</span>
        </div>
        <div className="stat-card">
          <span className="stat-card__label">Avg Basket</span>
          <span className="stat-card__value">€16.45</span>
          <span className="stat-card__delta">Per order</span>
        </div>
      </div>

      {/* Tabs */}
      <div className="overview__tabs">
        <button
          type="button"
          className={`overview__tab ${tab === 'orders' ? 'overview__tab--active' : ''}`}
          onClick={() => setTab('orders')}
        >
          Orders
        </button>
        <button
          type="button"
          className={`overview__tab ${tab === 'products' ? 'overview__tab--active' : ''}`}
          onClick={() => setTab('products')}
        >
          Products
        </button>
        <button
          type="button"
          className={`overview__tab ${tab === 'customers' ? 'overview__tab--active' : ''}`}
          onClick={() => setTab('customers')}
        >
          Customers
        </button>
      </div>

      {/* Tab content */}
      {tab === 'orders' && (
        <>
          {/* Monthly orders line chart */}
          <div className="overview__chart-card overview__chart-card--full">
            <h3 className="overview__chart-title">Orders & Deliveries per Month</h3>
            <ResponsiveContainer width="100%" height={280}>
              <LineChart data={monthlyData} margin={{ top: 10, right: 20, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                <XAxis dataKey="month" tick={{ fontSize: 12 }} />
                <YAxis tick={{ fontSize: 12 }} />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="orders" stroke="#e8b04b" strokeWidth={2} dot={false} />
                <Line type="monotone" dataKey="deliveries" stroke="#3a2e22" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Peak hours + Fulfillment donut */}
          <div className="overview__row">
            <div className="overview__chart-card overview__chart-card--half">
              <h3 className="overview__chart-title">Peak Hours</h3>
              <ResponsiveContainer width="100%" height={240}>
                <BarChart data={peakHours} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="hour" tick={{ fontSize: 11 }} />
                  <YAxis tick={{ fontSize: 12 }} />
                  <Tooltip />
                  <Bar dataKey="orders" fill="#e8b04b" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="overview__chart-card overview__chart-card--half">
              <h3 className="overview__chart-title">Order Fulfillment</h3>
              <ResponsiveContainer width="100%" height={240}>
                <PieChart>
                  <Pie
                    data={fulfillment}
                    cx="50%"
                    cy="50%"
                    innerRadius={55}
                    outerRadius={85}
                    paddingAngle={3}
                    dataKey="value"
                    label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
                    labelLine={false}
                  >
                    {fulfillment.map((entry) => (
                      <Cell key={entry.name} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>
        </>
      )}

      {tab === 'products' && (
        <div className="overview__chart-card overview__chart-card--full">
          <h3 className="overview__chart-title">Sells per Product</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={productSales} layout="vertical" margin={{ top: 10, right: 20, left: 100, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis type="number" tick={{ fontSize: 12 }} />
              <YAxis type="category" dataKey="name" tick={{ fontSize: 12 }} width={100} />
              <Tooltip />
              <Bar dataKey="sold" fill="#e8b04b" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {tab === 'customers' && (
        <div className="overview__chart-card overview__chart-card--full">
          <h3 className="overview__chart-title">Top 5 Customers</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={topCustomers} layout="vertical" margin={{ top: 10, right: 20, left: 80, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
              <XAxis type="number" tick={{ fontSize: 12 }} />
              <YAxis type="category" dataKey="name" tick={{ fontSize: 12 }} width={80} />
              <Tooltip />
              <Bar dataKey="orders" fill="#3a2e22" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}
