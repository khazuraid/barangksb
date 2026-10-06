import { initializeApp, getApps, getApp } from 'firebase/app';
import { 
  getAuth, 
  signInWithPopup, 
  GoogleAuthProvider, 
  onAuthStateChanged, 
  signOut,
  User,
  Auth
} from 'firebase/auth';
import firebaseConfig from '@/firebase-applet-config.json';

// Initialize Firebase safely
let authInstance: Auth | null = null;

function getAuthSafe(): Auth | null {
  if (typeof window === 'undefined') return null;
  if (!authInstance) {
    try {
      const app = getApps().length > 0 ? getApp() : initializeApp(firebaseConfig);
      authInstance = getAuth(app);
    } catch (e) {
      console.warn('Firebase initialization deferred/restricted:', e);
      return null;
    }
  }
  return authInstance;
}

// Provider with Google Workspace Scopes
const provider = new GoogleAuthProvider();
provider.addScope('https://www.googleapis.com/auth/spreadsheets');
provider.addScope('https://www.googleapis.com/auth/drive.file');

// Flag to indicate if we are in the middle of a sign-in flow
let isSigningIn = false;
// Cache the access token in memory ONLY (never in localStorage/sessionStorage)
let cachedAccessToken: string | null = null;

export const SCOPES = [
  'https://www.googleapis.com/auth/spreadsheets',
  'https://www.googleapis.com/auth/drive.file'
];

/**
 * Initialize auth state listener with guarded error callback
 */
export const initAuth = (
  onAuthSuccess?: (user: User, token: string) => void,
  onAuthFailure?: () => void
) => {
  try {
    const auth = getAuthSafe();
    if (!auth) {
      if (onAuthFailure) onAuthFailure();
      return () => {};
    }

    // Always provide onError callback to prevent uncaught window Script errors
    return onAuthStateChanged(
      auth,
      async (user: User | null) => {
        if (user) {
          if (cachedAccessToken) {
            if (onAuthSuccess) onAuthSuccess(user, cachedAccessToken);
          } else if (!isSigningIn) {
            cachedAccessToken = null;
            if (onAuthFailure) onAuthFailure();
          }
        } else {
          cachedAccessToken = null;
          if (onAuthFailure) onAuthFailure();
        }
      },
      (error) => {
        console.warn('Firebase onAuthStateChanged error (handled):', error);
        cachedAccessToken = null;
        if (onAuthFailure) onAuthFailure();
      }
    );
  } catch (err) {
    console.warn('initAuth caught exception:', err);
    if (onAuthFailure) onAuthFailure();
    return () => {};
  }
};

/**
 * Perform Google Sign-In with popup
 */
export const googleSignIn = async (): Promise<{ user: User; accessToken: string } | null> => {
  const auth = getAuthSafe();
  if (!auth) {
    throw new Error('Firebase Auth tidak tersedia di lingkungan ini.');
  }

  try {
    isSigningIn = true;
    const result = await signInWithPopup(auth, provider);
    const credential = GoogleAuthProvider.credentialFromResult(result);
    if (!credential?.accessToken) {
      throw new Error('Gagal mendapatkan token akses dari Google Auth');
    }

    cachedAccessToken = credential.accessToken;
    return { user: result.user, accessToken: cachedAccessToken };
  } catch (error: unknown) {
    console.error('Sign in error:', error);
    throw error;
  } finally {
    isSigningIn = false;
  }
};

/**
 * Retrieve current cached access token in memory
 */
export const getAccessToken = async (): Promise<string | null> => {
  return cachedAccessToken;
};

/**
 * Logout and clear token
 */
export const logout = async () => {
  try {
    const auth = getAuthSafe();
    if (auth) {
      await signOut(auth);
    }
  } catch (e) {
    console.warn('Sign out warning:', e);
  }
  cachedAccessToken = null;
};
