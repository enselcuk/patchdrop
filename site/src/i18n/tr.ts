import type { Dict } from './types'

const tr: Dict = {
  meta: {
    title: 'PatchDrop: ofis kablosundan dosya paylaşın',
    description:
      'Mac ve Windows arasında, kablolu ofis ağınız üzerinden ücretsiz ve açık kaynak dosya paylaşımı. İnternet ve hesap gerekmez. Beta.',
  },
  beta: {
    label: 'Beta',
    text: 'Ofisimizde Mac ve Windows’ta test edildi.',
    ask: 'Bir şey çalışmıyor mu?',
    link: 'GitHub’da bize yazın',
    dismiss: 'Bu notu gizle',
  },
  nav: {
    how: 'Nasıl çalışır',
    features: 'Özellikler',
    download: 'İndir',
    faq: 'SSS',
    github: 'GitHub',
    language: 'Dil',
    home: 'PatchDrop ana sayfa',
    main: 'Ana menü',
  },
  hero: {
    titleLead: 'Kablo üzerinden,',
    titleAccent: 'bir sürükle-bırak uzaklıkta.',
    lead: 'Ofis ağındaki Mac ve Windows bilgisayarlar arasında dosya paylaşmanın en kolay yolu. İnternet, bulut ya da hesap gerekmez.',
    downloadMac: 'macOS için indir',
    downloadWindows: 'Windows için indir',
    star: 'GitHub’da yıldızla',
    starsLabel: 'GitHub yıldızı',
    downloads: 'indirme',
    alsoWindows: 'Windows için de var',
    alsoMac: 'macOS için de var',
    otherNote: 'macOS ve Windows için',
    free: 'Ücretsiz ve açık kaynak',
    beta: 'beta',
    mobileNote: 'PatchDrop, Mac ve Windows’ta çalışır. İndirmek için bu sayfayı bilgisayarınızda açın.',
    sendToComputer: 'Bağlantıyı bilgisayarıma gönder',
    screenshotAlt:
      'Mac’te PatchDrop penceresi: cihaz listesinde ofisteki iki Mac, Muhasebe-Mac ve Tasarım-Mac, yanında dosyaları bırakma alanı.',
  },
  why: {
    eyebrow: 'Neden PatchDrop',
    title: 'Bilgisayarları kabloyla bağlı ofisler için.',
    paragraphs: [
      'Pek çok ofiste her masa kabloludur: bilgisayarlar ağa Ethernet kablosuyla bağlanır. Hız için harika, ama Wi-Fi düşünülerek tasarlanmış dosya paylaşım uygulamaları bu ağlarda çoğu zaman birbirini bulamıyor. Dosyalar da yan masaya hâlâ USB bellekle ya da e-postayla gidiyor.',
      'PatchDrop, bilgisayarınızdaki her kablolu ağ bağdaştırıcısında kendini duyurur. Böylece ofisteki bilgisayarlar birkaç saniye içinde birbirini bulur. Ağınız bunu da engelliyorsa bilgisayarı IP adresiyle ekleyebilir ya da ağı taratabilirsiniz.',
      'LocalSend’i çok seviyoruz; Wi-Fi’de harika bir seçenek. PatchDrop onun yerini almaya çalışmıyor. Bizi sürekli uğraştıran tek bir durumu çözüyor: kablolu ofis ağını.',
    ],
  },
  how: {
    title: 'Nasıl çalışır',
    sub: 'Üç adım. Hesap yok, ayar yok.',
    steps: [
      { title: 'İki bilgisayara da kurun', text: 'PatchDrop’u hem gönderecek hem de alacak bilgisayara indirin.' },
      { title: 'Dosyaları pencereye bırakın', text: 'Dosyaları ya da klasörlerin tamamını PatchDrop’a sürükleyin veya tıklayıp seçin.' },
      { title: 'Diğer bilgisayarı seçip gönderin', text: 'Listede kendiliğinden görünür. Karşı taraf kabul eder, dosyalar gelir.' },
    ],
  },
  features: {
    title: 'Özellikler',
    sub: 'Küçük bir uygulama, tek işi iyi yapar.',
    items: [
      { title: 'Kablo hızında', text: 'Dosyalar doğrudan ofis ağınızdan, kablonun izin verdiği en yüksek hızla gider.' },
      { title: 'Şifreli, doğrulanmış cihazlar', text: 'Her aktarım TLS 1.3 ile şifrelenir. Her bilgisayarın kendi sertifikası vardır; sertifika beklenenle eşleşmezse PatchDrop göndermez.' },
      { title: 'İnternet ve bulut yok', text: 'Hiçbir şey ağınızın dışına çıkmaz. İnternet kesikken bile çalışır.' },
      { title: 'Klasörün tamamı', text: 'Bir klasör gönderin, alt klasörleriyle birlikte olduğu gibi ulaşsın.' },
      { title: 'Mac + Windows', text: 'Intel ve Apple Silicon Mac’ler, Windows 10 ve 11. Hangi yönde olursa olsun.' },
      { title: 'Ücretsiz, MIT, reklamsız', text: 'MIT lisanslı açık kaynak. Reklam yok, hesap yok, ücretli sürüm yok.' },
    ],
  },
  download: {
    title: 'PatchDrop’u indirin',
    version: 'Sürüm',
    mac: {
      name: 'macOS',
      req: 'macOS 12 ve üstü · Intel ve Apple Silicon',
      primary: '.dmg indir',
      secondary: 'Universal .zip',
    },
    windows: {
      name: 'Windows',
      req: 'Windows 10 ve 11 · 64 bit',
      primary: 'Kurulum dosyasını indir',
      secondary: 'Taşınabilir .zip',
    },
    or: 'ya da',
    macTitle: 'Mac’te ilk açılış',
    macIntro: 'PatchDrop henüz Apple tarafından onaylanmadığı için macOS ilk açılışta engeller. İzin vermek için:',
    macSteps: [
      '.dmg dosyasını açın ve PatchDrop’u Uygulamalar klasörüne sürükleyin.',
      'PatchDrop’u bir kez açın. macOS engellediğinde Bitti’ye tıklayın.',
      'Sistem Ayarları → Gizlilik ve Güvenlik bölümüne gidin.',
      'Aşağı kaydırıp PatchDrop’un yanındaki Yine de Aç’a tıklayın ve onaylayın. Sorulursa yerel ağ erişimine izin verin.',
    ],
    macFigure: 'Gizlilik ve Güvenlik',
    macFigureText: '“PatchDrop”, Mac’inizi korumak için engellendi.',
    macFigureButton: 'Yine de Aç',
    winTitle: 'Windows’ta ilk açılış',
    winIntro: 'Kurulum dosyası yeni ve henüz imzalı olmadığı için SmartScreen uyarı verebilir:',
    winSteps: [
      '“Windows bilgisayarınızı korudu” ekranında Ek bilgi’ye tıklayın.',
      'Yine de çalıştır’a tıklayın ve kurulumu tamamlayın.',
      'Windows güvenlik duvarını sorarsa PatchDrop’a özel ağlarda izin verin.',
    ],
    releaseNotes: 'Sürüm notları',
    checksums: 'SHA-256 sağlama değerleri',
  },
  share: {
    title: 'PatchDrop’u paylaşın',
    lead: 'Mac ve Windows bilgisayarları aynı kabloya bağlı bir ofis mi biliyorsunuz? Bağlantıyı gönderin.',
    text: 'Kablolu ofis ağında Mac ve Windows arasında dosya paylaşmak için ücretsiz ve açık kaynak bir uygulama: PatchDrop',
    shareOn: 'Paylaş:',
    linkedin: 'LinkedIn',
    x: 'X',
    facebook: 'Facebook',
    whatsapp: 'WhatsApp',
    urlLabel: 'Sayfa bağlantısı',
    copy: 'Bağlantıyı kopyala',
    copied: 'Bağlantı kopyalandı',
    native: 'Paylaş…',
    selectHint: 'Bağlantıyı seçip kopyalayın',
  },
  faq: {
    title: 'Sık sorulan sorular',
    items: [
      {
        q: 'Gerçekten ücretsiz mi?',
        a: 'Evet. PatchDrop, MIT lisanslı, ücretsiz ve açık kaynak bir uygulama. Reklam, ücretli sürüm ya da hesap yok.',
      },
      {
        q: 'İnternet gerekiyor mu?',
        a: 'Hayır. Bilgisayarlar yerel ağınız üzerinden doğrudan konuşur. Ofiste internet kesilse de çalışmaya devam eder.',
      },
      {
        q: 'Güvenli mi?',
        a: 'Aktarımlar TLS 1.3 ile şifrelenir ve her bilgisayarın kendi sertifikası vardır. PatchDrop göndermeden önce karşı bilgisayarın sertifikasının beklenenle eşleştiğini kontrol eder; eşleşmezse durur. Karşıdaki kişi “Accept” düğmesine tıklamadan hiçbir dosya alınmaz; tabii sizden gelenleri her zaman kabul etmeyi seçmediyse.',
      },
      {
        q: 'Bilgisayarlar birbirini görmüyor. Ne yapmalıyım?',
        a: 'Önce güvenlik duvarına bakın ve iki bilgisayarda da PatchDrop’a izin verin. Windows’ta Ethernet bağlantısını Özel ağ olarak ayarlayın (Ayarlar → Ağ ve İnternet → Ethernet). Hâlâ görünmüyorsa diğer bilgisayarın IP adresini uygulamadaki “Add by IP” alanına yazın ya da “Scan network” düğmesine tıklayın (uygulama şimdilik İngilizce). PatchDrop bilgisayarları bulmak için UDP 47820, aktarım için TCP 47821 portunu kullanır; ikisinin de açık olması gerekir.',
      },
      {
        q: 'Neden doğrudan LocalSend kullanmayayım?',
        a: 'LocalSend harika bir uygulama; Wi-Fi’de biz de onu öneririz. Ancak kablolu ofis ağlarında cihaz bulma özelliği diğer bilgisayarları çoğu zaman bulamıyor. PatchDrop tam da bu durum için yapıldı. Ağınızda hangisi iyi çalışıyorsa onu kullanın.',
      },
      {
        q: 'Gelen dosyalar nereye kaydediliyor?',
        a: 'İndirilenler klasöründeki PatchDrop klasörüne. Uygulamanın ayarlarından başka bir klasör seçebilirsiniz. Aynı adlı dosyaların üzerine asla yazılmaz.',
      },
      {
        q: 'Beta ne demek?',
        a: 'PatchDrop’u ofisimizde birkaç iş arkadaşımız Mac ve Windows’ta test etti, ama henüz her cihazda ve her sistem sürümünde denenmedi. Bir sorun çıkarsa lütfen GitHub’da bildirin; çok işimize yarar.',
      },
      {
        q: 'Bu sayfa beni takip ediyor mu?',
        a: 'Yalnızca anonim ziyaret sayıları, GoatCounter ile. Çerez yok, kişisel veri yok.',
      },
    ],
  },
  madeBy: {
    eyebrow: 'Kim yaptı',
    title: 'Geliştiren: Selçuk Engin · Engin Sistem',
    bio: 'Selçuk Engin, 2003’ten bu yana B2B ve işletmelere özel yazılım sistemleri geliştiren İstanbullu yazılım şirketi Engin Sistem’in kurucusu. PatchDrop, kendi ofisinin kablolu ağındaki bir soruna çözüm olarak doğdu. Artık herkes için ücretsiz.',
    photoAlt: 'Selçuk Engin',
    personLinks: 'Selçuk’u takip edin',
    brandLinks: 'Engin Sistem’i takip edin',
    website: 'enginsistem.com’u ziyaret edin',
  },
  footer: {
    copyright: '© 2026 Engin Tasarım Ltd. Şti.',
    github: 'GitHub',
    bug: 'Hata bildir',
    releases: 'Sürümler',
    license: 'MIT Lisansı',
    links: 'Proje bağlantıları',
  },
  notFound: {
    title: 'Sayfa bulunamadı',
    description: 'Aradığınız sayfa yok ya da taşınmış. PatchDrop ana sayfasına dönebilirsiniz.',
    back: 'Ana sayfaya dön',
  },
}

export default tr
