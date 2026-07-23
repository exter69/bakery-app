import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ImageUpload from './ImageUpload';

// Mock the uploadImage function
vi.mock('../../api/seller', () => ({
  uploadImage: vi.fn(),
}));

import { uploadImage } from '../../api/seller';

const mockUploadImage = vi.mocked(uploadImage);

describe('ImageUpload', () => {
  const onChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the label when provided', () => {
    render(<ImageUpload value="" onChange={onChange} label="Photo du produit" />);
    expect(screen.getByText('Photo du produit')).toBeInTheDocument();
  });

  it('shows placeholder when no value', () => {
    render(<ImageUpload value="" onChange={onChange} />);
    expect(screen.getByText('Cliquez ou glissez une image')).toBeInTheDocument();
    expect(screen.getByText('JPEG, PNG ou WebP - max 5 Mo')).toBeInTheDocument();
  });

  it('shows preview image when value is set', () => {
    render(<ImageUpload value="/uploads/products/test.jpg" onChange={onChange} />);
    const img = screen.getByAltText('Apercu') as HTMLImageElement;
    expect(img).toBeInTheDocument();
    expect(img.src).toContain('/uploads/products/test.jpg');
  });

  it('rejects files with invalid type', async () => {
    render(<ImageUpload value="" onChange={onChange} />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    const file = new File(['content'], 'doc.pdf', { type: 'application/pdf' });
    // Use fireEvent to bypass accept attribute filtering
    Object.defineProperty(input, 'files', { value: [file] });
    await act(async () => {
      input.dispatchEvent(new Event('change', { bubbles: true }));
    });

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Type de fichier invalide');
    });
    expect(mockUploadImage).not.toHaveBeenCalled();
    expect(onChange).not.toHaveBeenCalled();
  });

  it('rejects files exceeding 5 MB', async () => {
    render(<ImageUpload value="" onChange={onChange} />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    // Create a file larger than 5 MB
    const largeContent = new Uint8Array(6 * 1024 * 1024);
    const file = new File([largeContent], 'big.jpg', { type: 'image/jpeg' });
    await userEvent.upload(input, file);

    expect(screen.getByRole('alert')).toHaveTextContent('Fichier trop volumineux');
    expect(mockUploadImage).not.toHaveBeenCalled();
    expect(onChange).not.toHaveBeenCalled();
  });

  it('calls uploadImage and onChange on valid file', async () => {
    mockUploadImage.mockResolvedValue('/uploads/products/abc.jpg');

    render(<ImageUpload value="" onChange={onChange} type="products" />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    const file = new File(['imagedata'], 'photo.jpg', { type: 'image/jpeg' });
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(mockUploadImage).toHaveBeenCalledWith(file, 'products');
      expect(onChange).toHaveBeenCalledWith('/uploads/products/abc.jpg');
    });
  });

  it('shows error when upload fails', async () => {
    mockUploadImage.mockRejectedValue(new Error('Upload failed'));

    render(<ImageUpload value="" onChange={onChange} />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    const file = new File(['imagedata'], 'photo.png', { type: 'image/png' });
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Upload failed');
    });
    expect(onChange).not.toHaveBeenCalled();
  });

  it('shows loading state during upload', async () => {
    // Never-resolving promise to keep uploading state
    mockUploadImage.mockImplementation(() => new Promise(() => {}));

    render(<ImageUpload value="" onChange={onChange} />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    const file = new File(['imagedata'], 'photo.webp', { type: 'image/webp' });
    await userEvent.upload(input, file);

    expect(screen.getByText('Telechargement...')).toBeInTheDocument();
  });

  it('uses the type prop for upload category', async () => {
    mockUploadImage.mockResolvedValue('/uploads/bakeries/abc.jpg');

    render(<ImageUpload value="" onChange={onChange} type="bakeries" />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;

    const file = new File(['imagedata'], 'shop.jpg', { type: 'image/jpeg' });
    await userEvent.upload(input, file);

    await waitFor(() => {
      expect(mockUploadImage).toHaveBeenCalledWith(file, 'bakeries');
    });
  });
});
