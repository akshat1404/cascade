/**
 * offlineDb.ts
 * Thin promise-based wrapper around IndexedDB for persisting unsaved
 * document state when the user is offline.
 *
 * DB layout:
 *   name    : cascade_offline  (version 1)
 *   store   : pending_saves    (keyPath: 'docId')
 *   record  : { docId, title, content, savedAt }
 */

const DB_NAME    = 'cascade_offline';
const STORE_NAME = 'pending_saves';
const DB_VERSION = 1;

/** Opens (and, on first run, creates) the IndexedDB database. */
function openDb(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const req = indexedDB.open(DB_NAME, DB_VERSION);

		req.onupgradeneeded = () => {
			req.result.createObjectStore(STORE_NAME, { keyPath: 'docId' });
		};
		req.onsuccess = () => resolve(req.result);
		req.onerror  = () => reject(req.error);
	});
}

export interface OfflineSave {
	docId   : string;
	title   : string;
	content : unknown;   // TipTap JSON (JSONContent)
	savedAt : number;    // Unix ms timestamp
}

/** Upsert an offline save record for the given document. */
export async function saveOfflineDoc(
	docId   : string,
	title   : string,
	content : unknown,
): Promise<void> {
	const db = await openDb();
	return new Promise((resolve, reject) => {
		const tx  = db.transaction(STORE_NAME, 'readwrite');
		tx.objectStore(STORE_NAME).put({ docId, title, content, savedAt: Date.now() });
		tx.oncomplete = () => { db.close(); resolve(); };
		tx.onerror    = () => { db.close(); reject(tx.error); };
	});
}

/** Retrieve the offline save record for the given document, or undefined. */
export async function getOfflineDoc(docId: string): Promise<OfflineSave | undefined> {
	const db = await openDb();
	return new Promise((resolve, reject) => {
		const tx  = db.transaction(STORE_NAME, 'readonly');
		const req = tx.objectStore(STORE_NAME).get(docId);
		req.onsuccess = () => { db.close(); resolve(req.result as OfflineSave | undefined); };
		req.onerror   = () => { db.close(); reject(req.error); };
	});
}

/** Delete the offline save record once it has been synced to the server. */
export async function deleteOfflineDoc(docId: string): Promise<void> {
	const db = await openDb();
	return new Promise((resolve, reject) => {
		const tx = db.transaction(STORE_NAME, 'readwrite');
		tx.objectStore(STORE_NAME).delete(docId);
		tx.oncomplete = () => { db.close(); resolve(); };
		tx.onerror    = () => { db.close(); reject(tx.error); };
	});
}
