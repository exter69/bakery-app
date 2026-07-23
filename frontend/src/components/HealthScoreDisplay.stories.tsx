import type { Meta, StoryObj } from '@storybook/react';
import HealthScoreDisplay from './HealthScoreDisplay';

const meta: Meta<typeof HealthScoreDisplay> = {
  title: 'Components/HealthScoreDisplay',
  component: HealthScoreDisplay,
  argTypes: {
    score: { control: { type: 'range', min: 1, max: 5, step: 1 } },
  },
};
export default meta;
type Story = StoryObj<typeof HealthScoreDisplay>;

export const Score1: Story = { args: { score: 1 } };
export const Score2: Story = { args: { score: 2 } };
export const Score3: Story = { args: { score: 3 } };
export const Score4: Story = { args: { score: 4 } };
export const Score5: Story = { args: { score: 5 } };
