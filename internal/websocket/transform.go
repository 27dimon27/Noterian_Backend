package websocket

func TransformCursorPosition(cursorStarPos int, cursorEndPos int, operationType MessageType, operation any, blockID string, userID string) (int, int) {
	transformedStartPos := cursorStarPos
	transformedEndPos := cursorEndPos

	switch operationType {
	case MsgInsertChars:
		if insertOp, ok := operation.(*InsertCharsOperation); ok {
			if insertOp.BlockID == blockID {
				insertLen := len([]rune(insertOp.Char))
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
		if deleteCharsOp, ok := operation.(*DeleteCharsOperation); ok {
			if deleteCharsOp.BlockID == blockID {
				switch {
				case deleteCharsOp.EndPosition <= transformedStartPos:
					transformedStartPos -= (deleteCharsOp.EndPosition - deleteCharsOp.StartPosition)
					transformedEndPos -= (deleteCharsOp.EndPosition - deleteCharsOp.StartPosition)
				case deleteCharsOp.StartPosition < transformedStartPos && deleteCharsOp.EndPosition < transformedEndPos && deleteCharsOp.EndPosition > transformedStartPos:
					transformedStartPos = deleteCharsOp.StartPosition
					transformedEndPos -= deleteCharsOp.EndPosition - deleteCharsOp.StartPosition
				case deleteCharsOp.StartPosition >= transformedStartPos && deleteCharsOp.EndPosition <= transformedEndPos:
					transformedEndPos -= deleteCharsOp.EndPosition - deleteCharsOp.StartPosition
				case deleteCharsOp.StartPosition < transformedStartPos && deleteCharsOp.EndPosition > transformedEndPos:
					transformedStartPos = deleteCharsOp.StartPosition
					transformedEndPos = deleteCharsOp.StartPosition
				case deleteCharsOp.StartPosition > transformedStartPos && deleteCharsOp.EndPosition > transformedEndPos && deleteCharsOp.StartPosition < transformedEndPos:
					transformedEndPos = deleteCharsOp.StartPosition
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
