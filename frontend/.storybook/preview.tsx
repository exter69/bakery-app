import type { Preview } from '@storybook/react';
import { I18nProvider } from '../src/i18n/I18nContext';
import { ThemeProvider } from '../src/theme/ThemeContext';
import '../src/index.css';

const preview: Preview = {
  decorators: [
    (Story) => (
      <ThemeProvider>
        <I18nProvider>
          <Story />
        </I18nProvider>
      </ThemeProvider>
    ),
  ],
  parameters: {
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
  },
};

export default preview;
