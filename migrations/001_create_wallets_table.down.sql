-- Drop indexes
DROP INDEX IF EXISTS idx_wallets_created_at;
DROP INDEX IF EXISTS idx_wallets_network;
DROP INDEX IF EXISTS idx_wallets_public_key;

-- Drop wallets table
DROP TABLE IF EXISTS wallets;

