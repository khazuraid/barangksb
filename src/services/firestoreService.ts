import { 
  getFirestore, 
  collection, 
  doc, 
  setDoc, 
  deleteDoc, 
  onSnapshot, 
  getDocFromServer,
  getDocs,
  Firestore 
} from 'firebase/firestore';
import { initializeApp, getApps, getApp } from 'firebase/app';
import firebaseConfig from '@/firebase-applet-config.json';
import { InventoryItem, StockTransaction, DEFAULT_CATEGORIES } from '../types/inventory';

let dbInstance: Firestore | null = null;

export function getDb(): Firestore | null {
  if (typeof window === 'undefined') return null;
  if (!dbInstance) {
    try {
      const app = getApps().length > 0 ? getApp() : initializeApp(firebaseConfig);
      // Connect to the provisioned database ID
      const dbId = (firebaseConfig as any).firestoreDatabaseId;
      dbInstance = dbId ? getFirestore(app, dbId) : getFirestore(app);
    } catch (err) {
      console.warn('Firestore initialization deferred/restricted:', err);
      return null;
    }
  }
  return dbInstance;
}

export async function testConnection(): Promise<boolean> {
  const db = getDb();
  if (!db) return false;
  try {
    await getDocFromServer(doc(db, 'test', 'connection'));
    return true;
  } catch (error) {
    console.warn('Firestore test connection notice:', error);
    return false;
  }
}

/**
 * Real-time listener for Inventory Items in Firestore
 */
export function subscribeItems(onUpdate: (items: InventoryItem[]) => void): () => void {
  const db = getDb();
  if (!db) return () => {};

  try {
    const colRef = collection(db, 'inventory_items');
    return onSnapshot(
      colRef,
      (snapshot) => {
        if (!snapshot.empty) {
          const items: InventoryItem[] = [];
          snapshot.forEach((docSnap) => {
            items.push(docSnap.data() as InventoryItem);
          });
          onUpdate(items);
        }
      },
      (err) => {
        console.warn('Firestore subscribeItems error:', err);
      }
    );
  } catch (e) {
    console.warn('Failed to subscribe to items:', e);
    return () => {};
  }
}

// Helper to sanitize objects for Firestore (removes undefined fields which Firestore rejects)
function cleanForFirestore<T>(obj: T): any {
  if (obj === null || obj === undefined) return null;
  if (Array.isArray(obj)) return obj.map(cleanForFirestore);
  if (typeof obj === 'object') {
    const cleaned: Record<string, any> = {};
    for (const [key, val] of Object.entries(obj)) {
      if (val !== undefined) {
        cleaned[key] = cleanForFirestore(val);
      }
    }
    return cleaned;
  }
  return obj;
}

/**
 * Save / Update Item in Firestore
 */
export async function saveItemToFirestore(item: InventoryItem): Promise<void> {
  const db = getDb();
  if (!db) return;
  try {
    const docRef = doc(db, 'inventory_items', item.id);
    await setDoc(docRef, cleanForFirestore(item), { merge: true });
  } catch (err) {
    console.warn('Failed to save item to Firestore (handled):', err);
  }
}

/**
 * Delete Item from Firestore
 */
export async function deleteItemFromFirestore(itemId: string): Promise<void> {
  const db = getDb();
  if (!db) return;
  try {
    const docRef = doc(db, 'inventory_items', itemId);
    await deleteDoc(docRef);
  } catch (err) {
    console.warn('Failed to delete item from Firestore (handled):', err);
  }
}

/**
 * Real-time listener for Stock-In Transactions
 */
export function subscribeTransactions(onUpdate: (txs: StockTransaction[]) => void): () => void {
  const db = getDb();
  if (!db) return () => {};

  try {
    const colRef = collection(db, 'stock_in_logs');
    return onSnapshot(
      colRef,
      (snapshot) => {
        if (!snapshot.empty) {
          const txs: StockTransaction[] = [];
          snapshot.forEach((docSnap) => {
            txs.push(docSnap.data() as StockTransaction);
          });
          // Sort by timestamp desc
          txs.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
          onUpdate(txs);
        }
      },
      (err) => {
        console.warn('Firestore subscribeTransactions error:', err);
      }
    );
  } catch (e) {
    console.warn('Failed to subscribe to transactions:', e);
    return () => {};
  }
}

/**
 * Save Transaction to Firestore
 */
export async function saveTransactionToFirestore(tx: StockTransaction): Promise<void> {
  const db = getDb();
  if (!db) return;
  try {
    const docRef = doc(db, 'stock_in_logs', tx.id);
    await setDoc(docRef, cleanForFirestore(tx));
  } catch (err) {
    console.warn('Failed to save transaction to Firestore (handled):', err);
  }
}

/**
 * Real-time listener for Categories
 */
export function subscribeCategories(onUpdate: (cats: string[]) => void): () => void {
  const db = getDb();
  if (!db) return () => {};

  try {
    const colRef = collection(db, 'categories');
    return onSnapshot(
      colRef,
      (snapshot) => {
        if (!snapshot.empty) {
          const cats: string[] = [];
          snapshot.forEach((docSnap) => {
            const data = docSnap.data();
            if (data.name) cats.push(data.name);
          });
          // Merge with DEFAULT_CATEGORIES
          const merged = Array.from(new Set([...DEFAULT_CATEGORIES, ...cats]));
          onUpdate(merged);
        }
      },
      (err) => {
        console.warn('Firestore subscribeCategories error:', err);
      }
    );
  } catch (e) {
    console.warn('Failed to subscribe to categories:', e);
    return () => {};
  }
}

/**
 * Add custom category to Firestore
 */
export async function saveCategoryToFirestore(name: string): Promise<void> {
  const db = getDb();
  if (!db) return;
  try {
    const catId = name.toLowerCase().replace(/[^a-z0-9]/g, '_');
    const docRef = doc(db, 'categories', catId);
    await setDoc(docRef, { id: catId, name });
  } catch (err) {
    console.error('Failed to save category to Firestore:', err);
  }
}

/**
 * Seed initial sample items if Firestore collection is empty
 */
export async function seedInitialFirestoreData(
  initialItems: InventoryItem[],
  initialTransactions: StockTransaction[]
): Promise<void> {
  const db = getDb();
  if (!db) return;

  try {
    const colRef = collection(db, 'inventory_items');
    const snap = await getDocs(colRef);
    if (snap.empty) {
      console.log('Seeding initial inventory to Firestore...');
      for (const item of initialItems) {
        await setDoc(doc(db, 'inventory_items', item.id), cleanForFirestore(item));
      }
      for (const tx of initialTransactions) {
        await setDoc(doc(db, 'stock_in_logs', tx.id), cleanForFirestore(tx));
      }
    }
  } catch (err) {
    console.warn('Seeding note:', err);
  }
}
