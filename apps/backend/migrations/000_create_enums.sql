DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_role') THEN
        CREATE TYPE project_role AS ENUM ('owner', 'member');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'join_request_status') THEN
        CREATE TYPE join_request_status AS ENUM ('pending', 'accepted', 'rejected');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'incident_priority') THEN
        CREATE TYPE incident_priority AS ENUM ('low', 'medium', 'high', 'critical');
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'incident_status') THEN
        CREATE TYPE incident_status AS ENUM ('open', 'in_progress', 'submitted', 'finished');
    END IF;
END $$;