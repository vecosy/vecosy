package vconf

import "context"

func (s *Server) GetFile(ctx context.Context, request *GetFileRequest) (*GetFileResponse, error) {
	file, err := s.repo.GetFile(request.AppName, request.AppVersion, request.FilePath)
	if err != nil {
		return nil, err
	}
	return &GetFileResponse{
		FileContent: file.Content,
	}, nil
}
