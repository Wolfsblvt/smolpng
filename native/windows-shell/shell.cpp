// SPDX-License-Identifier: MIT
// This native ABI bridge only displays a data-driven menu and launches the Go app.
// Compression, network, credentials, config evaluation and result writes never run
// on Explorer's menu-building path. No persistent server or managed runtime.
#define NOMINMAX
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <shobjidl.h>
#include <shlobj.h>
#include <shlwapi.h>
#include <wrl/client.h>
#include <atomic>
#include <memory>
#include <new>
#include <cwchar>
#include "core.hpp"
using Microsoft::WRL::ComPtr;
static HMODULE module_handle;
static std::atomic<long> live_objects {
    0
};
// Must match internal/shellmenu.ClassID and AppxManifest.xml.in.
static const CLSID MenuClass= {
    0x90bb1f36,0x193a,0x4d1a, {
        0x89,0x1b,0x91,0xdb,0x43,0xae,0x63,0xdc
    }
};
static std::wstring wide(std::u16string_view s) {
    static_assert(sizeof(wchar_t)==2);
    return std::wstring(reinterpret_cast<const wchar_t*>(s.data()),s.size());
}
static std::u16string narrow(std::wstring_view s) {
    return std::u16string(reinterpret_cast<const char16_t*>(s.data()),s.size());
}
struct Handle  {
    HANDLE value=INVALID_HANDLE_VALUE;
    ~Handle() {
        if(value!=INVALID_HANDLE_VALUE)CloseHandle(value);
    }
    Handle(const Handle&)=delete;
    Handle&operator=(const Handle&)=delete;
    explicit Handle(HANDLE h):value(h) {
    }
};
static std::wstring directory()  {
    std::vector<wchar_t> name(32768);
    DWORD n=GetModuleFileNameW(module_handle,name.data(),static_cast<DWORD>(name.size()));
    if(!n||n==name.size())throw std::runtime_error("module location");
    std::wstring p(name.data(),n);
    auto slash=p.find_last_of(L"\\/");
    if(slash==p.npos)throw std::runtime_error("module parent");
    return p.substr(0,slash);
}
static std::wstring user_root() {
    PWSTR p=nullptr;
    HRESULT hr=SHGetKnownFolderPath(FOLDERID_LocalAppData,0,nullptr,&p);
    if(FAILED(hr))throw std::runtime_error("user folder");
    std::wstring root(p);
    CoTaskMemFree(p);
    return root+L"\\smolpng";
}
static std::vector<std::uint8_t> read_menu()  {
    auto path=user_root()+L"\\shell-menu\\smolpng-shell-menu.ini";
    Handle f(CreateFileW(path.c_str(),GENERIC_READ,FILE_SHARE_READ|FILE_SHARE_WRITE|FILE_SHARE_DELETE,nullptr,OPEN_EXISTING,FILE_FLAG_OPEN_REPARSE_POINT,nullptr));
    if(f.value==INVALID_HANDLE_VALUE)throw std::runtime_error("menu not installed");
    BY_HANDLE_FILE_INFORMATION info {
    };
    if(!GetFileInformationByHandle(f.value,&info)||info.dwFileAttributes&(FILE_ATTRIBUTE_DIRECTORY|FILE_ATTRIBUTE_REPARSE_POINT)||info.nFileSizeHigh||info.nFileSizeLow>65536)throw std::runtime_error("invalid menu file");
    std::vector<std::uint8_t>b(info.nFileSizeLow);
    DWORD read=0;
    if(!ReadFile(f.value,b.data(),static_cast<DWORD>(b.size()),&read,nullptr)||read!=b.size())throw std::runtime_error("menu read");
    return b;
}
static smolpng::Menu load_menu() {
    auto m=smolpng::parse_menu(read_menu());
    if(_wcsicmp(wide(m.executable).c_str(),(directory()+L"\\smolpng.exe").c_str()))throw std::runtime_error("another installation owns this menu");
    return m;
}
static HRESULT selection(IShellItemArray* items,std::vector<std::u16string>* paths)  {
    if(!items)return E_INVALIDARG;
    DWORD count=0;
    HRESULT hr=items->GetCount(&count);
    if(FAILED(hr)||count==0||count>smolpng::MaxFiles)return E_INVALIDARG;
    for(DWORD i=0;i<count;++i) {
        ComPtr<IShellItem> item;
        hr=items->GetItemAt(i,&item);
        if(FAILED(hr))return hr;
        SFGAOF attributes=0;
        hr=item->GetAttributes(SFGAO_FILESYSTEM|SFGAO_FOLDER,&attributes);
        if(FAILED(hr)||!(attributes&SFGAO_FILESYSTEM)||attributes&SFGAO_FOLDER)return E_INVALIDARG;
        PWSTR p=nullptr;
        hr=item->GetDisplayName(SIGDN_FILESYSPATH,&p);
        if(FAILED(hr))return hr;
        std::wstring file(p);
        CoTaskMemFree(p);
        auto u=narrow(file);
        if(u.empty()||u.size()>32767||!smolpng::supported(u)||!smolpng::valid_utf16(u)||PathIsRelativeW(file.c_str()))return E_INVALIDARG;
        if(paths)paths->push_back(std::move(u));
    }
    return S_OK;
}
static HRESULT launch(const smolpng::Menu& menu,const smolpng::Action& action,IShellItemArray* items)  {
    auto current=load_menu();
    if(current.generation!=menu.generation)return HRESULT_FROM_WIN32(ERROR_REVISION_MISMATCH);
    std::vector<std::u16string>files;
    HRESULT hr=selection(items,&files);
    if(FAILED(hr))return hr;
    auto bytes=smolpng::request(menu,action,files);
    auto requests=user_root()+L"\\shell-requests";
    int err=SHCreateDirectoryExW(nullptr,requests.c_str(),nullptr);
    if(err!=ERROR_SUCCESS&&err!=ERROR_ALREADY_EXISTS&&err!=ERROR_FILE_EXISTS)return HRESULT_FROM_WIN32(err);
    DWORD attrib=GetFileAttributesW(requests.c_str());
    if(attrib==INVALID_FILE_ATTRIBUTES||attrib&FILE_ATTRIBUTE_REPARSE_POINT||!(attrib&FILE_ATTRIBUTE_DIRECTORY))return E_ACCESSDENIED;
    GUID uuid {
    };
    if(FAILED(CoCreateGuid(&uuid)))return E_FAIL;
    wchar_t guid[40] {
    };
    StringFromGUID2(uuid,guid,40);
    std::wstring hex;
    for(auto c:guid) {
        if((c>=L'0'&&c<=L'9')||(c>=L'a'&&c<=L'f'))hex+=c;
        else if(c>=L'A'&&c<=L'F')hex+=wchar_t(c+32);
    }
    if(hex.size()!=32)return E_FAIL;
    auto path=requests+L"\\smolpng-"+hex+L".request";
    {
        Handle f(CreateFileW(path.c_str(),GENERIC_WRITE,0,nullptr,CREATE_NEW,FILE_ATTRIBUTE_TEMPORARY,nullptr));
        if(f.value==INVALID_HANDLE_VALUE)return HRESULT_FROM_WIN32(GetLastError());
        DWORD written=0;
        bool ok=WriteFile(f.value,bytes.data(),static_cast<DWORD>(bytes.size()),&written,nullptr)&&written==bytes.size()&&FlushFileBuffers(f.value);
        if(!ok) {
            DWORD error=GetLastError();
            CloseHandle(f.value);
            f.value=INVALID_HANDLE_VALUE;
            DeleteFileW(path.c_str());
            return HRESULT_FROM_WIN32(error?error:ERROR_WRITE_FAULT);
        }
    }
    auto exe=directory()+(action.interactive?L"\\smolpng.exe":L"\\smolpngw.exe");
    auto command=wide(smolpng::quote(narrow(exe)))+L" _shell-run --request "+wide(smolpng::quote(narrow(path)));
    STARTUPINFOW startup {
    };
    startup.cb=sizeof(startup);
    PROCESS_INFORMATION process {
    };
    DWORD flags=CREATE_UNICODE_ENVIRONMENT|CREATE_NEW_PROCESS_GROUP|(action.interactive?CREATE_NEW_CONSOLE:CREATE_NO_WINDOW);
    if(!CreateProcessW(exe.c_str(),command.data(),nullptr,nullptr,FALSE,flags,nullptr,directory().c_str(),&startup,&process)) {
        DWORD error=GetLastError();
        DeleteFileW(path.c_str());
        return HRESULT_FROM_WIN32(error);
    }
    CloseHandle(process.hThread);
    CloseHandle(process.hProcess);
    return S_OK;
}
class Command;
class Enumerator final:public IEnumExplorerCommand  {
    std::atomic<ULONG>refs {
        1
    };
    std::shared_ptr<const smolpng::Menu> menu;
    std::size_t position=0;
    public:  explicit Enumerator(std::shared_ptr<const smolpng::Menu>m,std::size_t p=0):menu(std::move(m)),position(p) {
        ++live_objects;
    }
    ~Enumerator() {
        --live_objects;
    }
    IFACEMETHODIMP QueryInterface(REFIID iid,void**out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        if(iid==IID_IUnknown||iid==__uuidof(IEnumExplorerCommand)) {
            *out=static_cast<IEnumExplorerCommand*>(this);
            AddRef();
            return S_OK;
        }
        return E_NOINTERFACE;
    }
    IFACEMETHODIMP_(ULONG) AddRef()override {
        return ++refs;
    }
    IFACEMETHODIMP_(ULONG) Release()override {
        ULONG n=--refs;
        if(!n)delete this;
        return n;
    }
    IFACEMETHODIMP Next(ULONG count,IExplorerCommand**out,ULONG*fetched)override;
    IFACEMETHODIMP Skip(ULONG count)override {
        auto available=menu->actions.size()-position;
        position+=std::min<std::size_t>(count,available);
        return count<=available?S_OK:S_FALSE;
    }
    IFACEMETHODIMP Reset()override {
        position=0;
        return S_OK;
    }
    IFACEMETHODIMP Clone(IEnumExplorerCommand**out)override {
        if(!out)return E_POINTER;
        *out=new(std::nothrow)Enumerator(menu,position);
        return *out?S_OK:E_OUTOFMEMORY;
    }
};
class Command final:public IExplorerCommand  {
    std::atomic<ULONG>refs {
        1
    };
    std::shared_ptr<const smolpng::Menu> menu;
    int action_index=-1;
    const smolpng::Action&action()const {
        return menu->actions[action_index<0?0:static_cast<std::size_t>(action_index)];
    }
    bool submenu()const {
        return action_index<0&&menu&&menu->mode==u"submenu";
    }
    public:  Command() {
        ++live_objects;
        try {
            menu=std::make_shared<const smolpng::Menu>(load_menu());
        }
        catch(...) {
        }
    }
    Command(std::shared_ptr<const smolpng::Menu>m,int i):menu(std::move(m)),action_index(i) {
        ++live_objects;
    }
    ~Command() {
        --live_objects;
    }
    IFACEMETHODIMP QueryInterface(REFIID iid,void**out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        if(iid==IID_IUnknown||iid==__uuidof(IExplorerCommand)) {
            *out=static_cast<IExplorerCommand*>(this);
            AddRef();
            return S_OK;
        }
        return E_NOINTERFACE;
    }
    IFACEMETHODIMP_(ULONG) AddRef()override {
        return ++refs;
    }
    IFACEMETHODIMP_(ULONG) Release()override {
        ULONG n=--refs;
        if(!n)delete this;
        return n;
    }
    IFACEMETHODIMP GetTitle(IShellItemArray*,PWSTR*out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        if(!menu)return E_FAIL;
        try {
            return SHStrDupW(submenu()?L"smolpng":wide(action().label).c_str(),out);
        }
        catch(...) {
            return E_OUTOFMEMORY;
        }
    }
    IFACEMETHODIMP GetIcon(IShellItemArray*,PWSTR*out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        return E_NOTIMPL;
    }
    IFACEMETHODIMP GetToolTip(IShellItemArray*,PWSTR*out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        return SHStrDupW(L"Make selected images smaller locally. Originals are kept unless the action explicitly says replace.",out);
    }
    IFACEMETHODIMP GetCanonicalName(GUID*out)override {
        if(!out)return E_POINTER;
        *out=MenuClass;
        if(menu&&action_index>=0) {
            std::uint32_t hash=2166136261;
            for(auto c:action().id) {
                hash^=c;
                hash*=16777619;
            }
            out->Data1^=hash;
        }
        return S_OK;
    }
    IFACEMETHODIMP GetState(IShellItemArray*items,BOOL allow_slow,EXPCMDSTATE*out)override {
        if(!out)return E_POINTER;
        *out=ECS_HIDDEN;
        if(!menu)return S_OK;
        try {
            DWORD n=0;
            if(items&&SUCCEEDED(items->GetCount(&n))&&n>16&&!allow_slow)return E_PENDING;
            if(SUCCEEDED(selection(items,nullptr)))*out=ECS_ENABLED;
            return S_OK;
        }
        catch(...) {
            return E_FAIL;
        }
    }
    IFACEMETHODIMP Invoke(IShellItemArray*items,IBindCtx*)override {
        if(!menu||submenu())return E_UNEXPECTED;
        try {
            return launch(*menu,action(),items);
        }
        catch(const std::bad_alloc&) {
            return E_OUTOFMEMORY;
        }
        catch(...) {
            return E_FAIL;
        }
    }
    IFACEMETHODIMP GetFlags(EXPCMDFLAGS*out)override {
        if(!out)return E_POINTER;
        *out=submenu()?ECF_HASSUBCOMMANDS:ECF_DEFAULT;
        return S_OK;
    }
    IFACEMETHODIMP EnumSubCommands(IEnumExplorerCommand**out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        if(!submenu())return E_NOTIMPL;
        *out=new(std::nothrow)Enumerator(menu);
        return *out?S_OK:E_OUTOFMEMORY;
    }
};
HRESULT Enumerator::Next(ULONG count,IExplorerCommand**out,ULONG*fetched) {
    if(!out||(!fetched&&count!=1))return E_POINTER;
    ULONG n=0;
    for(;n<count&&position<menu->actions.size();++n,++position) {
        out[n]=new(std::nothrow)Command(menu,static_cast<int>(position));
        if(!out[n]) {
            for(ULONG j=0;j<n;++j)out[j]->Release();
            if(fetched)*fetched=0;
            return E_OUTOFMEMORY;
        }
    }
    if(fetched)*fetched=n;
    return n==count?S_OK:S_FALSE;
}
class Factory final:public IClassFactory  {
    std::atomic<ULONG>refs {
        1
    };
    public:  Factory() {
        ++live_objects;
    }
    ~Factory() {
        --live_objects;
    }
    IFACEMETHODIMP QueryInterface(REFIID iid,void**out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        if(iid==IID_IUnknown||iid==IID_IClassFactory) {
            *out=static_cast<IClassFactory*>(this);
            AddRef();
            return S_OK;
        }
        return E_NOINTERFACE;
    }
    IFACEMETHODIMP_(ULONG) AddRef()override {
        return ++refs;
    }
    IFACEMETHODIMP_(ULONG) Release()override {
        ULONG n=--refs;
        if(!n)delete this;
        return n;
    }
    IFACEMETHODIMP CreateInstance(IUnknown*outer,REFIID iid,void**out)override {
        if(!out)return E_POINTER;
        *out=nullptr;
        if(outer)return CLASS_E_NOAGGREGATION;
        auto p=new(std::nothrow)Command();
        if(!p)return E_OUTOFMEMORY;
        HRESULT hr=p->QueryInterface(iid,out);
        p->Release();
        return hr;
    }
    IFACEMETHODIMP LockServer(BOOL lock)override {
        if(lock)++live_objects;
        else --live_objects;
        return S_OK;
    }
};
extern "C" HRESULT __stdcall DllGetClassObject(REFCLSID clsid,REFIID iid,void**out) {
    if(!out)return E_POINTER;
    *out=nullptr;
    if(clsid!=MenuClass)return CLASS_E_CLASSNOTAVAILABLE;
    auto f=new(std::nothrow)Factory();
    if(!f)return E_OUTOFMEMORY;
    HRESULT hr=f->QueryInterface(iid,out);
    f->Release();
    return hr;
}
extern "C" HRESULT __stdcall DllCanUnloadNow() {
    return live_objects.load()==0?S_OK:S_FALSE;
}
BOOL WINAPI DllMain(HINSTANCE instance,DWORD reason,LPVOID) {
    if(reason==DLL_PROCESS_ATTACH) {
        module_handle=instance;
        DisableThreadLibraryCalls(instance);
    }
    return TRUE;
}
