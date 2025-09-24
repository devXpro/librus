# News Delivery Fix Documentation

## Problem Description

Previously, when multiple users from the same school received news/announcements, only the first user would get the news. This happened because:

1. News generate the same ID for all users in the same school (based on title + content + date)
2. The system stored only one news record per ID in the database
3. The `user_message_status` table tracked which messages were sent to which users
4. Since news had the same ID, the system thought all users already received the news after sending it to the first user

## Solution

Created a separate tracking system for news delivery:

### New Database Structure

1. **New Collection**: `user_news_status`
   - Tracks which news were sent to which users
   - Separate from regular message tracking
   - Allows the same news ID to be sent to multiple users

2. **New Model**: `UserNewsStatus`
   ```go
   type UserNewsStatus struct {
       Id             string    `bson:"_id"`              // Auto-generated ID
       TelegramUserID string    `bson:"telegram_user_id"` // Reference to TelegramUser._id
       NewsID         string    `bson:"news_id"`          // Reference to Message._id (for news only)
       SentAt         time.Time `bson:"sent_at"`          // When news was sent
   }
   ```

### New Functions

1. **News-specific tracking functions**:
   - `IsNewsSentToUser(telegramUserID, newsID string) bool`
   - `MarkNewsAsSent(telegramUserID, newsID string) error`

2. **Type-aware wrapper functions**:
   - `IsMessageSentToUserByType(telegramUserID, messageID string, messageType model.MessageType) bool`
   - `MarkMessageAsSentByType(telegramUserID, messageID string, messageType model.MessageType) error`

### Logic Changes

1. **Message Type Detection**: 
   - Regular messages (`MsgTypeMessage`) use the original `user_message_status` table
   - News/notifications (`MsgTypeNotification`, `MsgTypeNews`) use the new `user_news_status` table

2. **Updated Processing**:
   - `periodic_processing.go`: Uses type-aware functions for checking and marking messages
   - `menu.go`: Uses type-aware functions for manual message checking
   - `GenerateId()`: Added support for `MsgTypeNews` type

## Benefits

1. **No Duplication**: News are still stored once in the database (no duplication)
2. **Proper Delivery**: Each user gets news independently of other users
3. **Translation Caching**: Translation cache still works efficiently (based on text content, not message ID)
4. **Backward Compatibility**: Regular messages continue to work exactly as before
5. **Clean Separation**: News and messages have separate tracking systems

## Testing

Created comprehensive tests in `mongo/db_news_test.go`:
- Tests news status tracking functions
- Tests type-aware message handling
- Tests message ID generation consistency
- Verifies that identical news generate the same ID
- Verifies that different news generate different IDs

## Migration

No database migration is required:
- Existing data continues to work
- New collections are created automatically when first used
- Old message tracking remains intact

## Files Modified

1. `model/user.go` - Added `UserNewsStatus` struct
2. `mongo/db.go` - Added news tracking functions and type-aware wrappers
3. `model/message.go` - Updated `GenerateId()` to handle `MsgTypeNews`
4. `telegram/periodic_processing.go` - Updated to use type-aware functions
5. `telegram/handler/callback/menu.go` - Updated to use type-aware functions
6. `mongo/db_news_test.go` - Added comprehensive tests

## Usage

The fix is automatic and requires no configuration changes. The system will:
1. Automatically detect message types
2. Use appropriate tracking tables
3. Ensure all users receive their news
4. Maintain efficient translation caching
