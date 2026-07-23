import { useState, useRef, useCallback } from 'react';
import { uploadImage } from '../../api/seller';
import './ImageUpload.css';

const MAX_SIZE = 5 * 1024 * 1024; // 5 MB
const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

interface ImageUploadProps {
  value: string;
  onChange: (url: string) => void;
  label?: string;
  type?: 'products' | 'bakeries';
}

export default function ImageUpload({ value, onChange, label, type = 'products' }: ImageUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(async (file: File) => {
    setError(null);

    if (!ALLOWED_TYPES.includes(file.type)) {
      setError('Type de fichier invalide. Formats acceptes : JPEG, PNG, WebP.');
      return;
    }

    if (file.size > MAX_SIZE) {
      setError('Fichier trop volumineux. Taille maximale : 5 Mo.');
      return;
    }

    setUploading(true);
    try {
      const url = await uploadImage(file, type);
      onChange(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Echec du telechargement.');
    } finally {
      setUploading(false);
    }
  }, [onChange, type]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFile(file);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) {
      handleFile(file);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = () => {
    setDragOver(false);
  };

  const openPicker = () => {
    inputRef.current?.click();
  };

  return (
    <div className="image-upload">
      {label && <label className="image-upload__label">{label}</label>}

      <div
        className={`image-upload__dropzone ${dragOver ? 'image-upload__dropzone--active' : ''} ${uploading ? 'image-upload__dropzone--uploading' : ''}`}
        onClick={openPicker}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        role="button"
        tabIndex={0}
        aria-label={label || 'Telecharger une image'}
        onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') openPicker(); }}
      >
        {value && !uploading && (
          <img
            src={value}
            alt="Apercu"
            className="image-upload__preview"
          />
        )}

        {!value && !uploading && (
          <div className="image-upload__placeholder">
            <svg className="image-upload__icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="17 8 12 3 7 8" />
              <line x1="12" y1="3" x2="12" y2="15" />
            </svg>
            <span className="image-upload__text">Cliquez ou glissez une image</span>
            <span className="image-upload__hint">JPEG, PNG ou WebP - max 5 Mo</span>
          </div>
        )}

        {uploading && (
          <div className="image-upload__loading">
            <div className="image-upload__spinner" aria-label="Telechargement en cours" />
            <span>Telechargement...</span>
          </div>
        )}
      </div>

      {error && <p className="image-upload__error" role="alert">{error}</p>}

      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        onChange={handleChange}
        className="image-upload__input"
        aria-hidden="true"
        tabIndex={-1}
      />
    </div>
  );
}
