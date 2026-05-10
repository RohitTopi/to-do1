package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
)

type Todo struct {
	ID        int    `json:"id"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func apiURL(path string) string {
	return "http://localhost:4000" + path
}

func fetchTodos() ([]Todo, error) {
	resp, err := http.Get(apiURL("/api/todos"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var todos []Todo
	if err := json.Unmarshal(b, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func postTodo(body string) error {
	payload := map[string]string{"body": body}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(apiURL("/api/todos"), "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func patchTodoComplete(id string) error {
	req, err := http.NewRequest(http.MethodPatch, apiURL("/api/todos/"+id), nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func deleteTodo(id string) error {
	req, err := http.NewRequest(http.MethodDelete, apiURL("/api/todos/"+id), nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func main() {

	// application
	app := widgets.NewQApplication(len(os.Args), os.Args)

	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Todos (Qt UI)")
	window.SetMinimumSize2(400, 300)

	central := widgets.NewQWidget(nil, 0)
	vbox := widgets.NewQVBoxLayout()
	central.SetLayout(vbox)

	list := widgets.NewQListWidget(nil)
	vbox.AddWidget(list, 0, 0)

	h := widgets.NewQHBoxLayout()
	input := widgets.NewQLineEdit(nil)
	addBtn := widgets.NewQPushButton2("Add", nil)
	h.AddWidget(input, 1, 0)
	h.AddWidget(addBtn, 0, 0)
	vbox.AddLayout(h, 0)

	btnComplete := widgets.NewQPushButton2("Complete", nil)
	btnDelete := widgets.NewQPushButton2("Delete", nil)
	row := widgets.NewQHBoxLayout()
	row.AddWidget(btnComplete, 0, 0)
	row.AddWidget(btnDelete, 0, 0)
	vbox.AddLayout(row, 0)

	window.SetCentralWidget(central)
	window.Show()

	updateList := func() {
		list.Clear()
		todos, err := fetchTodos()
		if err != nil {
			list.AddItem("Error fetching TODOs: " + err.Error())
			return
		}
		for _, t := range todos {
			status := ""
			if t.Completed {
				status = " (done)"
			}
			list.AddItem(fmt.Sprintf("%d: %s%s", t.ID, t.Body, status))
		}
	}

	// wire add button
	addBtn.ConnectClicked(func(checked bool) {
		body := strings.TrimSpace(input.Text())
		if body == "" {
			return
		}
		if err := postTodo(body); err != nil {
			list.AddItem("Add error: " + err.Error())
			return
		}
		input.SetText("")
		updateList()
	})

	// wire complete button
	btnComplete.ConnectClicked(func(checked bool) {
		item := list.CurrentItem()
		if item == nil {
			return
		}
		text := item.Text()
		parts := strings.SplitN(text, ":", 2)
		if len(parts) < 1 {
			return
		}
		id := strings.TrimSpace(parts[0])
		if err := patchTodoComplete(id); err != nil {
			list.AddItem("Complete error: " + err.Error())
			return
		}
		updateList()
	})

	// wire delete button
	btnDelete.ConnectClicked(func(checked bool) {
		item := list.CurrentItem()
		if item == nil {
			return
		}
		text := item.Text()
		parts := strings.SplitN(text, ":", 2)
		if len(parts) < 1 {
			return
		}
		id := strings.TrimSpace(parts[0])
		if err := deleteTodo(id); err != nil {
			list.AddItem("Delete error: " + err.Error())
			return
		}
		updateList()
	})

	// initial load
	updateList()

	// this is to ensure app exits cleanly
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)
	app.Exec()
}
