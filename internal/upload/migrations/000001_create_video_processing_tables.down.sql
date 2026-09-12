-- 1. Remover Triggers
DROP TRIGGER IF EXISTS update_video_processings_updated_at ON video_processings;
DROP TRIGGER IF EXISTS update_video_batches_updated_at ON video_batches;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 2. Remover Índices
DROP INDEX IF EXISTS idx_video_processings_batch_id;
DROP INDEX IF EXISTS idx_video_processings_user_id;
DROP INDEX IF EXISTS idx_video_processings_status;

-- 3. Remover Tabelas (respeitando a restrição de FK)
DROP TABLE IF EXISTS video_processings;
DROP TABLE IF EXISTS video_batches;

-- 4. Remover Tipo Customizado ENUM
DROP TYPE IF EXISTS video_process_status;
