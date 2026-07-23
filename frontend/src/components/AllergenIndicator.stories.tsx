import type { Meta, StoryObj } from '@storybook/react';
import { AllergenIndicator } from './AllergenIndicator';

const meta: Meta<typeof AllergenIndicator> = {
  title: 'Components/AllergenIndicator',
  component: AllergenIndicator,
  argTypes: {
    allergens: { control: 'object' },
    productName: { control: 'text' },
  },
};
export default meta;
type Story = StoryObj<typeof AllergenIndicator>;

export const MultipleAllergens: Story = {
  args: {
    allergens: ['gluten', 'dairy', 'eggs'],
    productName: 'Croissant au beurre',
    onOpenModal: () => console.log('modal opened'),
  },
};

export const SingleAllergen: Story = {
  args: {
    allergens: ['nuts'],
    productName: 'Pain aux noix',
    onOpenModal: () => console.log('modal opened'),
  },
};
