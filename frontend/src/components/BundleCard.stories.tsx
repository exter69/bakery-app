import type { Meta, StoryObj } from '@storybook/react';
import { BundleCard } from './BundleCard';
import type { Bundle } from '../types/bundle';

const baseBundle: Bundle = {
  id: 'bundle-1',
  bakeryId: 'bakery-1',
  bakeryName: 'Boulangerie du Coin',
  bakeryLatitude: 48.8566,
  bakeryLongitude: 2.3522,
  name: 'Panier du soir',
  type: 'compose',
  photoUrl: 'https://images.unsplash.com/photo-1509440159596-0249088772ff?w=400',
  description: 'Assortiment de viennoiseries et pains de la journee',
  estimatedValue: 1200,
  originalPrice: 1000,
  discountedPrice: 450,
  quantityTotal: 5,
  quantityRemaining: 3,
  pickupStartTime: '18:00',
  pickupEndTime: '19:30',
  publishedDate: '2025-07-28',
  expiresAt: '2025-07-28T19:30:00Z',
  status: 'published',
  items: [
    { description: 'Croissant', quantity: 2 },
    { description: 'Pain au chocolat', quantity: 1 },
    { description: 'Baguette tradition', quantity: 1 },
  ],
  createdAt: '2025-07-28T06:00:00Z',
};

const meta: Meta<typeof BundleCard> = {
  title: 'Components/BundleCard',
  component: BundleCard,
  argTypes: {
    reserveLoading: { control: 'boolean' },
  },
};
export default meta;
type Story = StoryObj<typeof BundleCard>;

export const ComposeBundle: Story = {
  args: {
    bundle: baseBundle,
    userLatitude: 48.86,
    userLongitude: 2.35,
    onReserve: (id: string) => console.log('reserve', id),
    reserveLoading: false,
  },
};

export const SurpriseBundle: Story = {
  args: {
    bundle: {
      ...baseBundle,
      id: 'bundle-2',
      name: 'Panier Surprise',
      type: 'surprise',
      items: [],
      estimatedValue: 1500,
      description: '3 a 5 produits varies de la journee',
    },
    onReserve: (id: string) => console.log('reserve', id),
    reserveLoading: false,
  },
};

export const SoldOut: Story = {
  args: {
    bundle: {
      ...baseBundle,
      id: 'bundle-3',
      status: 'sold_out',
      quantityRemaining: 0,
    } satisfies Bundle,
    onReserve: (id: string) => console.log('reserve', id),
    reserveLoading: false,
  },
};
