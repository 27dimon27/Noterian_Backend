package websocket

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

func TransformCursorPosition(cursor CursorPosition, operationType MessageType, operation any, blockID string, userID string) (int, int) {
	transformedStartPos := cursor.StartPosition
	transformedEndPos := cursor.EndPosition

	switch operationType {
	case MsgInsertChars:
		if insertOp, ok := operation.(*InsertCharsOperation); ok {
			insertLen := len([]rune(insertOp.Char))

			if insertOp.UserID == userID {
				transformedStartPos += insertLen
				transformedEndPos = transformedStartPos
				return transformedStartPos, transformedEndPos
			}

			if insertOp.BlockID == blockID {
				switch {
				case insertOp.Position < transformedStartPos:
					transformedStartPos += insertLen
					transformedEndPos += insertLen
				case insertOp.Position >= transformedStartPos && insertOp.Position < transformedEndPos:
					transformedEndPos += insertLen
				}
			}
		}

	case MsgDeleteChars:
		if deleteOp, ok := operation.(*DeleteCharsOperation); ok {
			if deleteOp.UserID == userID {
				if transformedStartPos != transformedEndPos {
					transformedEndPos = transformedStartPos
					return transformedStartPos, transformedEndPos
				}

				transformedStartPos--
				transformedEndPos--
				return transformedStartPos, transformedEndPos
			}

			if deleteOp.BlockID == blockID {
				switch {
				case deleteOp.EndPosition <= transformedStartPos:
					transformedStartPos -= (deleteOp.EndPosition - deleteOp.StartPosition)
					transformedEndPos -= (deleteOp.EndPosition - deleteOp.StartPosition)
				case deleteOp.StartPosition < transformedStartPos && deleteOp.EndPosition < transformedEndPos && deleteOp.EndPosition > transformedStartPos:
					transformedStartPos = deleteOp.StartPosition
					transformedEndPos -= deleteOp.EndPosition - deleteOp.StartPosition
				case deleteOp.StartPosition >= transformedStartPos && deleteOp.EndPosition <= transformedEndPos:
					transformedEndPos -= deleteOp.EndPosition - deleteOp.StartPosition
				case deleteOp.StartPosition < transformedStartPos && deleteOp.EndPosition > transformedEndPos:
					transformedStartPos = deleteOp.StartPosition
					transformedEndPos = deleteOp.StartPosition
				case deleteOp.StartPosition > transformedStartPos && deleteOp.EndPosition > transformedEndPos && deleteOp.StartPosition < transformedEndPos:
					transformedEndPos = deleteOp.StartPosition
				}
			}
		}
	}

	if transformedEndPos < transformedStartPos {
		transformedStartPos = transformedEndPos
	}
	if transformedStartPos < 0 {
		transformedStartPos = 0
	}
	if transformedEndPos < 0 {
		transformedEndPos = 0
	}

	return transformedStartPos, transformedEndPos
}

func TransformCursorPositionAfterBlockOperation(cursor CursorPosition, operationType MessageType, operation any, authorCursor CursorPosition, block *models.Block) CursorPosition {
	switch operationType {
	case MsgCreateBlock:
		if createBlockOp, ok := operation.(*CreateBlockOperation); ok {
			if authorCursor.UserID == cursor.UserID {
				cursor.StartPosition = 0
				cursor.EndPosition = 0
				cursor.BlockID = createBlockOp.BlockID
				return cursor
			}

			if authorCursor.BlockID != cursor.BlockID {
				return cursor
			}

			if authorCursor.EndPosition < cursor.StartPosition {
				cursor.StartPosition = 0
				cursor.EndPosition = 0
				cursor.BlockID = createBlockOp.BlockID
				return cursor
			} else if authorCursor.StartPosition > cursor.StartPosition && authorCursor.EndPosition < cursor.EndPosition {
				cursor.EndPosition = authorCursor.StartPosition
				cursor.BlockID = createBlockOp.BlockID
				return cursor
			}
		}

	case MsgDeleteBlock:
		if _, ok := operation.(*DeleteBlockOperation); ok {
			if authorCursor.BlockID == cursor.BlockID {
				if block != nil {
					cursor.StartPosition = len(block.Content)
					cursor.EndPosition = len(block.Content)
					cursor.BlockID = block.ID.String()
				} else {
					cursor.StartPosition = 0
					cursor.EndPosition = 0
					cursor.BlockID = ""
				}
				return cursor
			}
		}
	}

	return cursor
}
