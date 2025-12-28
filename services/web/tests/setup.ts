import { GlobalRegistrator } from '@happy-dom/global-registrator';

// Register happy-dom globals FIRST (window, document, etc.)
GlobalRegistrator.register();

// Import jest-dom matchers AFTER happy-dom is registered
import '@testing-library/jest-dom';
