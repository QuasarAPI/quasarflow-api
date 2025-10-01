-- Create wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    public_key VARCHAR(56) NOT NULL UNIQUE,
    encrypted_key TEXT NOT NULL,
    network VARCHAR(10) NOT NULL CHECK (network IN ('testnet', 'mainnet')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on public_key for faster lookups
CREATE INDEX IF NOT EXISTS idx_wallets_public_key ON wallets(public_key);

-- Create index on network for filtering
CREATE INDEX IF NOT EXISTS idx_wallets_network ON wallets(network);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_wallets_created_at ON wallets(created_at DESC);

-- Add comment to table
COMMENT ON TABLE wallets IS 'Stores Stellar wallet information with encrypted private keys';
COMMENT ON COLUMN wallets.id IS 'Unique identifier for the wallet';
COMMENT ON COLUMN wallets.public_key IS 'Stellar public key (starts with G, 56 characters)';
COMMENT ON COLUMN wallets.encrypted_key IS 'AES-256-GCM encrypted private key (base64 encoded)';
COMMENT ON COLUMN wallets.network IS 'Stellar network (testnet or mainnet)';
COMMENT ON COLUMN wallets.created_at IS 'Timestamp when the wallet was created';
COMMENT ON COLUMN wallets.updated_at IS 'Timestamp when the wallet was last updated';

