CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Criar o tipo ENUM para os status caso ele não exista
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'video_process_status') THEN
        CREATE TYPE video_process_status AS ENUM (
            'PENDING', 
            'PROCESSING', 
            'COMPLETED', 
            'FAILED'
        );
    END IF;
END $$;

-- 2. Criar a tabela de agrupamento de lotes (Batches)
CREATE TABLE IF NOT EXISTS video_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 3. Criar a tabela principal de processamento de vídeos
CREATE TABLE IF NOT EXISTS video_processings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id UUID REFERENCES video_batches(id) ON DELETE SET NULL,
    status video_process_status NOT NULL DEFAULT 'PENDING',
    
    -- Dados do arquivo original
    name VARCHAR(255) NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    
    -- Dados pós-processamento
    output_path VARCHAR(512),
    
    -- Resiliência e erros
    error_message TEXT,
    
    -- Auditoria (padrão de compatibilidade com Go/GORM soft delete)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 4. Criar índices performáticos e seguros
CREATE INDEX IF NOT EXISTS idx_video_processings_status 
    ON video_processings(status) 
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_video_processings_batch_id 
    ON video_processings(batch_id) 
    WHERE batch_id IS NOT NULL AND deleted_at IS NULL;

-- 5. Função padrão para atualização do updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 6. Triggers de atualização automática de timestamp
CREATE TRIGGER update_video_batches_updated_at
    BEFORE UPDATE ON video_batches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_video_processings_updated_at
    BEFORE UPDATE ON video_processings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
