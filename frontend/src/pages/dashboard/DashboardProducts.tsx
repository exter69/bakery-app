import { useState, useEffect, useCallback } from 'react';
import { fetchMyBakery, fetchProducts, createProduct, updateProduct, deleteProduct } from '../../api/seller';
import type { Product } from '../../types/bakery';
import './Dashboard.css';

interface ProductForm {
  name: string;
  description: string;
  price: string;
  category: string;
  photoUrl: string;
}

const emptyForm: ProductForm = { name: '', description: '', price: '', category: '', photoUrl: '' };

export default function DashboardProducts() {
  const [bakeryId, setBakeryId] = useState<string | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // Modal state
  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ProductForm>(emptyForm);
  const [submitting, setSubmitting] = useState(false);

  const loadProducts = useCallback(async (bId: string) => {
    try {
      const prods = await fetchProducts(bId);
      setProducts(prods);
    } catch {
      setMsg({ type: 'error', text: 'Failed to load products.' });
    }
  }, []);

  useEffect(() => {
    fetchMyBakery()
      .then((b) => {
        if (b) {
          setBakeryId(b.id);
          return loadProducts(b.id);
        }
      })
      .catch(() => setMsg({ type: 'error', text: 'Failed to load bakery.' }))
      .finally(() => setLoading(false));
  }, [loadProducts]);

  const openAdd = () => {
    setEditingId(null);
    setForm(emptyForm);
    setShowModal(true);
  };

  const openEdit = (p: Product) => {
    setEditingId(p.id);
    setForm({
      name: p.name,
      description: p.description,
      price: (p.price / 100).toFixed(2),
      category: p.category,
      photoUrl: p.photoUrl,
    });
    setShowModal(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!bakeryId) return;
    setSubmitting(true);
    setMsg(null);
    try {
      const priceInCents = Math.round(parseFloat(form.price) * 100);
      if (editingId) {
        await updateProduct(editingId, {
          name: form.name,
          description: form.description,
          price: priceInCents,
          category: form.category,
          photoUrl: form.photoUrl,
        });
        setMsg({ type: 'success', text: 'Product updated.' });
      } else {
        await createProduct(bakeryId, {
          name: form.name,
          description: form.description,
          price: priceInCents,
          category: form.category,
          photoUrl: form.photoUrl,
        });
        setMsg({ type: 'success', text: 'Product created.' });
      }
      setShowModal(false);
      await loadProducts(bakeryId);
    } catch {
      setMsg({ type: 'error', text: 'Failed to save product.' });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!bakeryId || !window.confirm('Delete this product? This cannot be undone.')) return;
    try {
      await deleteProduct(id);
      setProducts((prev) => prev.filter((p) => p.id !== id));
      setMsg({ type: 'success', text: 'Product deleted.' });
    } catch {
      setMsg({ type: 'error', text: 'Failed to delete product.' });
    }
  };

  const handleToggleAvailability = async (p: Product) => {
    try {
      const updated = await updateProduct(p.id, { isAvailable: !p.isAvailable });
      setProducts((prev) => prev.map((x) => (x.id === updated.id ? updated : x)));
    } catch {
      setMsg({ type: 'error', text: 'Failed to update availability.' });
    }
  };

  const formatPrice = (cents: number) => `€${(cents / 100).toFixed(2)}`;

  if (loading) return <div className="dash-loading">Loading products…</div>;

  if (!bakeryId) {
    return (
      <div className="dash-empty">
        <h1 className="dash-page__title">Products</h1>
        <p style={{ marginTop: '1rem' }}>No bakery found.</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="dash-page__title">Products</h1>
      <p className="dash-page__subtitle">Manage your product catalog.</p>

      {msg && <div className={`dash-msg dash-msg--${msg.type}`}>{msg.text}</div>}

      <div className="dash-toolbar">
        <div className="dash-toolbar__spacer" />
        <button className="dash-btn dash-btn--primary" onClick={openAdd}>+ Add Product</button>
      </div>

      {products.length === 0 ? (
        <div className="dash-empty">No products yet. Add your first product above.</div>
      ) : (
        <div className="dash-card" style={{ padding: 0, overflow: 'hidden' }}>
          <div className="dash-table-wrap">
            <table className="dash-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Category</th>
                  <th>Price</th>
                  <th>Available</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {products.map((p) => (
                  <tr key={p.id}>
                    <td>{p.name}</td>
                    <td>{p.category}</td>
                    <td>{formatPrice(p.price)}</td>
                    <td>
                      <label className="dash-toggle">
                        <input
                          type="checkbox"
                          checked={p.isAvailable}
                          onChange={() => handleToggleAvailability(p)}
                        />
                        <span className="dash-toggle__slider" />
                      </label>
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button className="dash-btn dash-btn--secondary dash-btn--sm" onClick={() => openEdit(p)}>Edit</button>
                        <button className="dash-btn dash-btn--danger dash-btn--sm" onClick={() => handleDelete(p.id)}>Delete</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Modal */}
      {showModal && (
        <div className="dash-modal-overlay" onClick={() => setShowModal(false)}>
          <div className="dash-modal" onClick={(e) => e.stopPropagation()}>
            <h2 className="dash-modal__title">{editingId ? 'Edit Product' : 'Add Product'}</h2>
            <form className="dash-form" onSubmit={handleSubmit}>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-name">Name</label>
                <input id="prod-name" className="dash-form__input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-desc">Description</label>
                <textarea id="prod-desc" className="dash-form__textarea" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} rows={2} />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-price">Price (€)</label>
                <input id="prod-price" className="dash-form__input" type="number" step="0.01" min="0" value={form.price} onChange={(e) => setForm({ ...form, price: e.target.value })} required />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-category">Category</label>
                <input id="prod-category" className="dash-form__input" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} required />
              </div>
              <div className="dash-form__field">
                <label className="dash-form__label" htmlFor="prod-photo">Photo URL</label>
                <input id="prod-photo" className="dash-form__input" value={form.photoUrl} onChange={(e) => setForm({ ...form, photoUrl: e.target.value })} placeholder="https://..." />
              </div>
              <div className="dash-modal__actions">
                <button type="button" className="dash-btn dash-btn--secondary" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit" className="dash-btn dash-btn--primary" disabled={submitting}>
                  {submitting ? 'Saving…' : editingId ? 'Update' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
