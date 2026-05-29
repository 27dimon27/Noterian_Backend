package websocket

func TransformCursorPosition(cursorStarPos int, cursorEndPos int, operationType MessageType, operation any, blockID string, userID string) (int, int) {
	transformedStartPos := cursorStarPos
	transformedEndPos := cursorEndPos

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
