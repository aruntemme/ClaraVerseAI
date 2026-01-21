import { Turnstile, type TurnstileInstance } from '@marsidev/react-turnstile';
import { useRef, forwardRef, useImperativeHandle } from 'react';

interface TurnstileWidgetProps {
  onSuccess: (token: string) => void;
  onError?: () => void;
  onExpire?: () => void;
}

export interface TurnstileWidgetRef {
  reset: () => void;
}

export const TurnstileWidget = forwardRef<TurnstileWidgetRef, TurnstileWidgetProps>(
  ({ onSuccess, onError, onExpire }, ref) => {
    const turnstileRef = useRef<TurnstileInstance | null>(null);

    useImperativeHandle(ref, () => ({
      reset: () => {
        turnstileRef.current?.reset();
      },
    }));

    const siteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY;

    if (!siteKey) {
      console.warn('Turnstile site key not configured. CAPTCHA will not be displayed.');
      return null;
    }

    return (
      <div
        className="turnstile-container"
        style={{ display: 'flex', justifyContent: 'center', marginBottom: '1rem' }}
      >
        <Turnstile
          ref={turnstileRef}
          siteKey={siteKey}
          onSuccess={onSuccess}
          onError={() => {
            console.error('Turnstile error occurred');
            onError?.();
          }}
          onExpire={() => {
            console.warn('Turnstile token expired');
            onExpire?.();
          }}
          options={{
            theme: 'light',
            size: 'normal',
          }}
        />
      </div>
    );
  }
);
