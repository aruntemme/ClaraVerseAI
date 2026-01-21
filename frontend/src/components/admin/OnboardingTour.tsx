import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { driver } from 'driver.js';
import 'driver.js/dist/driver.css';

const PROVIDERS_TOUR_KEY = 'admin_providers_tour_completed';
const MODELS_TOUR_KEY = 'admin_models_tour_completed';

export const useOnboardingTour = () => {
  const location = useLocation();

  useEffect(() => {
    // Providers page tour - independent
    const hasCompletedProvidersTour = localStorage.getItem(PROVIDERS_TOUR_KEY);
    if (!hasCompletedProvidersTour && location.pathname.includes('/admin/providers')) {
      const timer = setTimeout(() => {
        const driverObj = driver({
          showProgress: true,
          steps: [
            {
              element: '[data-tour="add-provider"]',
              popover: {
                title: 'Add Your First Provider',
                description: 'Click here to add an AI provider like OpenAI, Anthropic, or Google. You\'ll need to provide an API key.',
                side: 'bottom',
                align: 'start',
              },
            },
            {
              popover: {
                title: 'After Adding a Provider',
                description: 'Once you add a provider, navigate to the Models page to fetch and enable models.',
              },
            },
            {
              element: '[data-tour="models-link"]',
              popover: {
                title: 'Models Page',
                description: 'After adding a provider, click here to manage models. You\'ll be able to fetch models from your providers and toggle their visibility.',
                side: 'right',
                align: 'start',
              },
            },
          ],
          onDestroyStarted: () => {
            localStorage.setItem(PROVIDERS_TOUR_KEY, 'true');
            driverObj.destroy();
          },
        });

        driverObj.drive();
      }, 500);

      return () => clearTimeout(timer);
    }

    // Models page tour - independent
    const hasCompletedModelsTour = localStorage.getItem(MODELS_TOUR_KEY);
    if (!hasCompletedModelsTour && location.pathname.includes('/admin/models')) {
      const timer = setTimeout(() => {
        const driverObj = driver({
          showProgress: true,
          steps: [
            {
              popover: {
                title: 'Welcome to Models Management',
                description: 'Here you can manage all your AI models. Let\'s walk through the key features.',
              },
            },
            {
              element: '[data-tour="toggle-visibility"]',
              popover: {
                title: 'Toggle Model Visibility',
                description: 'Click the Visible/Hidden button to control which models users can see and use. Only visible models will appear in the chat interface.',
                side: 'left',
                align: 'start',
              },
            },
            {
              popover: {
                title: 'Expand Model Details',
                description: 'Click on any model row (the chevron icon or model name) to expand and see detailed settings including capabilities and aliases.',
              },
            },
            {
              element: '[data-tour="model-capabilities"]',
              popover: {
                title: 'Model Capabilities Explained',
                description: '🔧 **Tools**: Enable function calling - lets the model use external APIs and functions.\n\n👁️ **Vision**: Allows the model to understand and analyze images.\n\n📡 **Streaming**: Enables real-time response streaming for better UX.\n\n⚡ **Smart Router**: Advanced routing for optimized model selection.\n\nClick each to toggle on/off.',
                side: 'top',
                align: 'start',
              },
            },
            {
              element: '[data-tour="model-aliases"]',
              popover: {
                title: 'Model Aliases',
                description: 'Aliases let you create alternative configurations for the same model. For example, you can have one alias with vision enabled and another without. This allows fine-grained control over model capabilities for different use cases.',
                side: 'top',
                align: 'start',
              },
            },
            {
              popover: {
                title: 'You\'re All Set!',
                description: 'You can now start using your configured models. Remember to enable visibility for models you want users to access, and configure their capabilities based on your needs.',
              },
            },
          ],
          onDestroyStarted: () => {
            localStorage.setItem(MODELS_TOUR_KEY, 'true');
            driverObj.destroy();
          },
        });

        driverObj.drive();
      }, 800);

      return () => clearTimeout(timer);
    }
  }, [location.pathname]);
};

export const resetOnboarding = () => {
  localStorage.removeItem(PROVIDERS_TOUR_KEY);
  localStorage.removeItem(MODELS_TOUR_KEY);
};
