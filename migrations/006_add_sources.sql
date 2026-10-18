INSERT INTO sources (name, type, url, fetch_interval_mins) VALUES

-- Media
('Engadget',                 'rss', 'https://www.engadget.com/rss.xml',                                                                  60),

-- Lab blogs
('OpenAI Blog',              'rss', 'https://openai.com/news/rss.xml',                                                                   60),
('Google DeepMind Blog',     'rss', 'https://deepmind.google/blog/rss.xml',                                                             120),
-- Anthropic has no official RSS; this community-maintained feed scrapes anthropic.com/news hourly
('Anthropic News',           'rss', 'https://raw.githubusercontent.com/taobojlen/anthropic-rss-feed/main/anthropic_news_rss.xml',       120),
('GitHub Blog AI',           'rss', 'https://github.blog/ai-and-ml/feed/',                                                             120),

-- Research journals
('JMLR',                     'rss', 'https://www.jmlr.org/jmlr.xml',                                                                   360),
('IEEE Xplore AI',           'rss', 'https://ieeexplore.ieee.org/rss/TOC9234890.XML',                                                  360)

ON CONFLICT (name) DO NOTHING;
