-- Switch Reddit sources from JSON API (403) to Atom feeds
UPDATE sources SET url = 'https://www.reddit.com/r/MachineLearning/.rss', fetch_interval_mins = 60
  WHERE name = 'Reddit r/MachineLearning';

UPDATE sources SET url = 'https://www.reddit.com/r/LocalLLaMA/.rss', fetch_interval_mins = 60
  WHERE name = 'Reddit r/LocalLLaMA';

-- Remove Reddit sources that exceed rate limits or are low-signal
DELETE FROM sources WHERE name IN ('Reddit r/artificial', 'Reddit r/Singularity');

-- Remove dead sources
DELETE FROM sources WHERE name IN ('Reuters Technology', 'Papers With Code');

-- Add new high-quality sources
INSERT INTO sources (name, type, url, fetch_interval_mins) VALUES
  ('Simon Willison',  'rss', 'https://simonwillison.net/atom/everything/', 60),
  ('Interconnects',   'rss', 'https://www.interconnects.ai/feed',          120)
ON CONFLICT (name) DO NOTHING;
