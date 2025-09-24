# Message Delivery System Refactoring Documentation

## Problem Description

The system had issues with message delivery for multiple users:

1. **News Issue**: Multiple users from the same school only received news once (only first user got them)
2. **Regular Messages Issue**: Messages with identical URLs were only delivered to the first account that processed them
3. **Complex Architecture**: Separate tracking tables for news vs regular messages
4. **Unnecessary Data**: `librus_login` field in messages was redundant

## Root Cause

The main issue was in `AddMessagesToDatabase()` function:
- It returned only **new** messages that were added to database
- If a message already existed (same ID), it returned an empty list
- This caused `periodic_processing.go` to skip sending messages to users of other accounts
- The `user_message_status` table was designed correctly but never got the chance to work

## Solution

Completely refactored the message delivery system:

### Key Changes

1. **Unified Message Delivery Tracking**
   - Replaced `user_message_status` and `user_news_status` with single `user_message_delivery` table
   - All message types (regular, notifications, news) now use the same tracking system

2. **Simplified Message Model**
   ```go
   type Message struct {
       Id             string      `bson:"_id"`                       // MD5 hash of URL or title+content+date
       Type           MessageType `bson:"type"`                      // message, notification, news
       Link           string      `bson:"link"`                      // URL to message in Librus
       Author         string      `bson:"author"`                    // Message author
       Title          string      `bson:"title"`                     // Message title
       Content        string      `bson:"content"`                   // Message content
       Date           time.Time   `bson:"date"`                      // Message date
       AttachmentsDir string      `bson:"attachments_dir,omitempty"` // Path to directory with attachments
       // Removed: LibrusLogin field (no longer needed)
   }
   ```

3. **Unified Delivery Tracking Model**
   ```go
   type UserMessageDelivery struct {
       Id             string    `bson:"_id"`              // Auto-generated ID
       TelegramUserID string    `bson:"telegram_user_id"` // Reference to TelegramUser._id
       MessageID      string    `bson:"message_id"`       // Reference to Message._id
       SentAt         time.Time `bson:"sent_at"`          // When message was sent
   }
   ```

4. **Simplified Functions**
   - `IsMessageSentToUser(telegramUserID, messageID string) bool` - works for all message types
   - `MarkMessageAsSent(telegramUserID, messageID string) error` - works for all message types
   - `AddMessagesToDatabase(messages []model.Message) error` - simplified, no return value

5. **Fixed Core Logic**
   - `AddMessagesToDatabase()` now only adds messages to database, doesn't filter what to process
   - `periodic_processing.go` always processes all messages from gRPC, regardless of whether they're new in DB
   - Message delivery tracking happens per-user, allowing same message to be sent to multiple users

## Benefits

1. **No Duplication**: Messages are still stored once in the database (no duplication)
2. **Proper Delivery**: Each user gets messages independently of other users
3. **Translation Caching**: Translation cache still works efficiently (based on text content, not message ID)
4. **Simplified Architecture**: Single tracking system for all message types
5. **Cleaner Code**: Removed unnecessary fields and functions
6. **Better Performance**: Simplified database operations

## Testing

Updated tests in `mongo/db_news_test.go`:
- Tests unified message delivery tracking functions
- Tests message ID generation consistency
- Verifies that identical messages generate the same ID
- Verifies that different messages generate different IDs
- Tests work with the new `user_message_delivery` collection

## Migration

No database migration is required:
- Existing data continues to work
- New collections are created automatically when first used
- Old message tracking remains intact

## Files Modified

1. `model/message.go` - Removed `LibrusLogin` field
2. `model/user.go` - Replaced `UserMessageStatus` and `UserNewsStatus` with unified `UserMessageDelivery`
3. `mongo/db.go` - Simplified and unified message delivery tracking functions
4. `telegram/periodic_processing.go` - Simplified logic, removed `addLibrusLoginToMessages()`
5. `telegram/handler/callback/menu.go` - Updated to use simplified functions
6. `telegram/handler/state/url.go` - Removed `LibrusLogin` assignment
7. `mongo/db_news_test.go` - Updated tests for unified system

## Database Collections

### Before
- `message` - messages with `librus_login` field
- `user_message_status` - tracking for regular messages
- `user_news_status` - tracking for news

### After
- `message` - messages without `librus_login` field
- `user_message_delivery` - unified tracking for all message types

## Usage

The fix is automatic and requires no configuration changes. The system will:
1. Process all messages from gRPC regardless of whether they're new in database
2. Use unified delivery tracking for all message types
3. Ensure all users receive their messages independently
4. Maintain efficient translation caching
5. Work with cleaner, simplified code architecture
