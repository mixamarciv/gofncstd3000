::получаем curpath:
@FOR /f %%i IN ("%0") DO SET curpath=%~dp0
::задаем основные переменные окружения
@CALL "%curpath%/set_path.bat"


@CLS

@echo ==== start ===========================================================
go get "github.com/satori/go.uuid"
go install
::go help gopath
@echo ==== end =============================================================

@PAUSE
