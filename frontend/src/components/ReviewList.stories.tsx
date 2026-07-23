import type { Meta, StoryObj } from '@storybook/react';
import ReviewList from './ReviewList';

const meta: Meta<typeof ReviewList> = {
  title: 'Components/ReviewList',
  component: ReviewList,
  parameters: {
    docs: {
      description: {
        component:
          'Displays reviews for a bakery. Fetches data from the API. ' +
          'In Storybook, the API call will fail gracefully showing the empty/loading state.',
      },
    },
  },
};
export default meta;
type Story = StoryObj<typeof ReviewList>;

/** With a bakery ID — will show loading then empty state without a running API. */
export const WithBakeryId: Story = {
  args: { bakeryId: 'bakery-123' },
};
