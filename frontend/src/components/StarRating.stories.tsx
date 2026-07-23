import type { Meta, StoryObj } from '@storybook/react';
import StarRating from './StarRating';

const meta: Meta<typeof StarRating> = {
  title: 'Components/StarRating',
  component: StarRating,
  argTypes: {
    rating: { control: { type: 'range', min: 0, max: 5, step: 0.5 } },
    size: { control: 'select', options: ['sm', 'md', 'lg'] },
  },
};
export default meta;
type Story = StoryObj<typeof StarRating>;

export const Display: Story = { args: { rating: 3.5, size: 'md' } };

export const Interactive: Story = {
  args: { rating: 0, size: 'lg', onChange: (r: number) => console.log(r) },
};

export const Small: Story = { args: { rating: 4, size: 'sm' } };

export const Empty: Story = { args: { rating: 0, size: 'md' } };

export const Full: Story = { args: { rating: 5, size: 'md' } };
