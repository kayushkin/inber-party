-- Rating System Database Schema
-- Add this to your database to support the bounty rating and reputation system

-- Create bounty_ratings table
CREATE TABLE IF NOT EXISTS bounty_ratings (
    id SERIAL PRIMARY KEY,
    bounty_id INTEGER NOT NULL,
    rater_id INTEGER NOT NULL,  -- Agent ID of the bounty creator who is rating
    rated_id INTEGER NOT NULL,  -- Agent ID of the bounty claimer who is being rated
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT DEFAULT '',
    categories JSONB DEFAULT '{}',  -- Breakdown: {"quality": 5, "timeliness": 4, "communication": 5}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (bounty_id) REFERENCES bounties(id) ON DELETE CASCADE,
    FOREIGN KEY (rater_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (rated_id) REFERENCES agents(id) ON DELETE CASCADE,
    
    -- Ensure one rating per bounty
    UNIQUE(bounty_id, rater_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_bounty_ratings_bounty_id ON bounty_ratings(bounty_id);
CREATE INDEX IF NOT EXISTS idx_bounty_ratings_rated_id ON bounty_ratings(rated_id);
CREATE INDEX IF NOT EXISTS idx_bounty_ratings_created_at ON bounty_ratings(created_at);

-- Add sample data for testing
INSERT INTO bounty_ratings (bounty_id, rater_id, rated_id, rating, comment, categories) VALUES 
(1, 1, 2, 5, 'Excellent work! Delivered ahead of schedule with high quality code.', '{"quality": 5, "timeliness": 5, "communication": 5}'),
(2, 1, 3, 4, 'Good work overall, minor issues with communication but solid delivery.', '{"quality": 4, "timeliness": 4, "communication": 3}'),
(3, 2, 1, 3, 'Average work, met basic requirements but could be improved.', '{"quality": 3, "timeliness": 3, "communication": 3}')
ON CONFLICT (bounty_id, rater_id) DO NOTHING;

-- Update bounties table to support better querying with ratings
ALTER TABLE bounties ADD COLUMN IF NOT EXISTS has_rating BOOLEAN DEFAULT FALSE;

-- Create a trigger to update has_rating flag when ratings are added
CREATE OR REPLACE FUNCTION update_bounty_rating_flag()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE bounties SET has_rating = TRUE WHERE id = NEW.bounty_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE bounties SET has_rating = (
            EXISTS (SELECT 1 FROM bounty_ratings WHERE bounty_id = OLD.bounty_id)
        ) WHERE id = OLD.bounty_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER IF NOT EXISTS trigger_bounty_rating_flag
AFTER INSERT OR DELETE ON bounty_ratings
FOR EACH ROW EXECUTE FUNCTION update_bounty_rating_flag();

-- Create a view for easy rating statistics
CREATE OR REPLACE VIEW agent_rating_stats AS
SELECT 
    a.id as agent_id,
    a.name as agent_name,
    COUNT(br.id) as total_ratings,
    AVG(br.rating) as average_rating,
    MIN(br.rating) as min_rating,
    MAX(br.rating) as max_rating,
    COUNT(CASE WHEN br.rating = 5 THEN 1 END) as five_star_count,
    COUNT(CASE WHEN br.rating = 4 THEN 1 END) as four_star_count,
    COUNT(CASE WHEN br.rating = 3 THEN 1 END) as three_star_count,
    COUNT(CASE WHEN br.rating = 2 THEN 1 END) as two_star_count,
    COUNT(CASE WHEN br.rating = 1 THEN 1 END) as one_star_count,
    COUNT(CASE WHEN br.created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as recent_ratings_30d,
    AVG(CASE WHEN br.created_at >= CURRENT_DATE - INTERVAL '30 days' THEN br.rating END) as recent_average_30d
FROM agents a
LEFT JOIN bounty_ratings br ON a.id = br.rated_id
GROUP BY a.id, a.name
ORDER BY average_rating DESC NULLS LAST, total_ratings DESC;

-- Create a view for bounty details with ratings
CREATE OR REPLACE VIEW bounty_details_with_ratings AS
SELECT 
    b.*,
    br.rating as rating_score,
    br.comment as rating_comment,
    br.created_at as rating_created_at,
    creator.name as creator_name,
    claimer.name as claimer_name,
    CASE 
        WHEN b.status = 'completed' AND br.id IS NULL AND b.creator_id IS NOT NULL THEN TRUE
        ELSE FALSE
    END as can_be_rated
FROM bounties b
LEFT JOIN bounty_ratings br ON b.id = br.bounty_id
LEFT JOIN agents creator ON b.creator_id = creator.id  
LEFT JOIN agents claimer ON b.claimer_id = claimer.id;

-- Add comment with instructions
COMMENT ON TABLE bounty_ratings IS 'Stores ratings (1-5 stars) given by bounty creators to claimers after work completion';
COMMENT ON COLUMN bounty_ratings.categories IS 'JSON object with category-specific ratings like {"quality": 5, "timeliness": 4, "communication": 5}';
COMMENT ON VIEW agent_rating_stats IS 'Aggregated rating statistics for each agent including distribution and recent ratings';
COMMENT ON VIEW bounty_details_with_ratings IS 'Complete bounty information with rating data and flags for rating eligibility';